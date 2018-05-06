package main

import (
    "fmt"
    "log"
    "crypto/sha256"
    "encoding/hex"
    "crypto/ecdsa"
    "strings"
    "crypto/rand"
    "bytes"
    "math/big"
    "encoding/gob"
    "crypto/elliptic"
)

const subsidy = 10

type Transaction struct {
    ID   []byte  // 该笔交易的交易ID
    Vin  []TXInput
    Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
    return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// A coinbase transaction has only one input.
func NewCoinbaseTX(to, data string) *Transaction {
    if data == "" {
            data = fmt.Sprintf("Reward to '%s'", to)
    }
        
    txin := TXInput{[]byte{}, -1, nil, []byte(data)}
    txout := NewTXOutput(subsidy, to)
    tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
    tx.ID = tx.Hash() 
    return &tx
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
    var encoded bytes.Buffer

    enc := gob.NewEncoder(&encoded)
    err := enc.Encode(tx)
    if err != nil {
            log.Panic(err)
    }
        
    return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
    var hash [32]byte

    txCopy := *tx
    txCopy.ID = []byte{}

    hash = sha256.Sum256(txCopy.Serialize())

    return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction){ // look at signing-scheme.png
    if tx.IsCoinbase() {           // in order to sign a transaction, we need to access the outputs referenced in the inputs of the transaction, thus we need the transactions that store these outputs.
            return
    }
    for _, vin := range tx.Vin {  // check Previous transaction is not correct 
            if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
                    log.Panic("ERROR: Previous transaction is not correct")
            }
    }
    txCopy := tx.TrimmedCopy()
    for inID, vin := range txCopy.Vin {  // inputs are signed separately
            prevTx := prevTXs[hex.EncodeToString(vin.Txid)] // get previous transaction
            txCopy.Vin[inID].Signature = nil   // Signature is set to nil (just a double-check)
            txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash // PubKey is set to the PubKeyHash of the referenced output.
            txCopy.ID = txCopy.Hash()           // The resulted hash is the data we’re going to sign
            txCopy.Vin[inID].PubKey = nil       // After getting the hash we should reset the PubKey field, so it doesn’t affect further iterations.
                        
            r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)// the central piece, privKey and the data we're going to sign
            if err != nil {
                    log.Panic(err)
            }
            signature := append(r.Bytes(), s.Bytes()...)
            tx.Vin[inID].Signature = signature
    }
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
    var lines []string

    lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

    for i, input := range tx.Vin {
            lines = append(lines, fmt.Sprintf("     Input %d:", i))
            lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
            lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
            lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
            lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
    }
    for i, output := range tx.Vout {
            lines = append(lines, fmt.Sprintf("     Output %d:", i))
            lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
            lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
    }
    return strings.Join(lines, "\n")
}

/*
Public key hashes stored in unlocked outputs. This identifies “sender” of a transaction.
Public key hashes stored in new, locked, outputs. This identifies “recipient” of a transaction.
Values of new outputs.
*/
// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
    var inputs []TXInput
    var outputs []TXOutput
    for _, vin := range tx.Vin {  // TXInput.Signature and TXInput.PubKey are set to nil.
            inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
    }
    for _, vout := range tx.Vout {  
            outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash}) 
    }
    txCopy := Transaction{tx.ID, inputs, outputs}
    return txCopy
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
    if tx.IsCoinbase() {
            return true
    }
        
    for _, vin := range tx.Vin {
            if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
                    log.Panic("ERROR: Previous transaction is not correct")
            }
    }
    txCopy := tx.TrimmedCopy()  // A trimmed copy will be signed, not a full transaction
    curve := elliptic.P256()    // used to generate key pairs
    for inID, vin := range tx.Vin {  // check signature in each input 
            prevTx := prevTXs[hex.EncodeToString(vin.Txid)]  
            txCopy.Vin[inID].Signature = nil  
            txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash 
            txCopy.ID = txCopy.Hash()
            txCopy.Vin[inID].PubKey = nil  // This piece is identical to the one in the Sign method, because during verification we need the same data what was signed.
// unpack values stored in TXInput.Signature and TXInput.PubKey, since a signature is a pair of numbers and a public key is a pair of coordinates. 
            r := big.Int{}
            s := big.Int{}
            sigLen := len(vin.Signature)
            r.SetBytes(vin.Signature[:(sigLen / 2)])
            s.SetBytes(vin.Signature[(sigLen / 2):])

            x := big.Int{}
            y := big.Int{}
            keyLen := len(vin.PubKey)
            x.SetBytes(vin.PubKey[:(keyLen / 2)])
            y.SetBytes(vin.PubKey[(keyLen / 2):])

            rawPubKey := ecdsa.PublicKey{curve, &x, &y} // private key sign, public key verify
            if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
                return false
            }
    }
    return true
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
    var inputs []TXInput
    var outputs []TXOutput

    wallets, err := NewWallets()
    if err != nil {
            log.Panic(err)
    }
    wallet := wallets.GetWallet(from)
    pubKeyHash := HashPubKey(wallet.PublicKey)
    acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

    if acc < amount {
            log.Panic("ERROR: Not enough funds")
    }
        
    // Build a list of input. 从能使用的output中构建input，比如tx0.Output 1，tx1.Output 0，tx3.Output 0等等
    for txid, outs := range validOutputs {  // range循环用在map时，txid as key， outs as value
            txID, err := hex.DecodeString(txid)
            if err != nil {
                    log.Panic(err)
            }
                                    
            for _, out := range outs {
                    input := TXInput{txID, out, nil, wallet.PublicKey}
                    inputs = append(inputs, input)
            }
    }
    // Build a list of outputs.                           create two outputs
    outputs = append(outputs, *NewTXOutput(amount, to)) // locked by receiver address 
    if acc > amount {
            outputs = append(outputs, *NewTXOutput(acc - amount, from)) // a change,  locked by sender address
    }

    tx := Transaction{nil, inputs, outputs}
    tx.ID = tx.Hash()
    bc.SignTransaction(&tx, wallet.PrivateKey)
    return &tx
}

