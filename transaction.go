package main

import (
    "fmt"
    "log"
    "crypto/sha256"
    "encoding/gob"
    "bytes"
    "encoding/hex"
    
)

const subsidy = 10

type Transaction struct {
    ID   []byte
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
    tx.SetID()
        
    return &tx
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
    var encoded bytes.Buffer
    var hash [32]byte

    enc := gob.NewEncoder(&encoded)
    err := enc.Encode(tx)
    if err != nil {
            log.Panic(err)
    }
    hash = sha256.Sum256(encoded.Bytes())
    tx.ID = hash[:]
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
        
    // Build a list of inputs
    for txid, outs := range validOutputs {
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
    tx.SetID()

    return &tx
}

