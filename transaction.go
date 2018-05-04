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

type TXOutput struct {
    Value       int
    PubKeyHash  []byte
}

func (out *TXOutput) Lock(address []byte) {
    pubKeyHash := Base58Decode(address)
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    out.PubKeyHash = pubKeyHash
}

// checks if provided public key hash was used to lock the output. 
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
    return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

type TXInput struct {
    Txid      []byte   // stores the ID of such transaction,
    Vout      int      // stores an index of an output in the transaction. 
    Signature []byte 
    PubKey    []byte
}

// checks that an input uses a specific key to unlock an output
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
    lockingHash := HashPubKey(in.PubKey)

    return bytes.Compare(lockingHash, pubKeyHash) == 0
}


// A coinbase transaction has only one input.
func NewCoinbaseTX(to, data string) *Transaction {
    if data == "" {
            data = fmt.Sprintf("Reward to '%s'", to)
    }
        
    txin := TXInput{[]byte{}, -1, nil, []byte(data)}
    txout := TXOutput{subsidy, to}
    tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
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

    acc, validOutputs := bc.FindSpendableOutputs(from, amount)

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
                    input := TXInput{txID, out, from}
                    inputs = append(inputs, input)
            }
    }
    // Build a list of outputs.                           create two outputs
    outputs = append(outputs, TXOutput{amount, to})  // locked by receiver address 
    if acc > amount {
            outputs = append(outputs, TXOutput{acc - amount, from}) // a change,  locked by sender address
    }

    tx := Transaction{nil, inputs, outputs}
    tx.SetID()

    return &tx
}

