package main

import (
    "fmt"
    "log"
    "crypto/sha256"
    "encoding/gob"
    "bytes"
)

const subsidy = 10

type Transaction struct {
    ID   []byte
    Vin  []TXInput
    Vout []TXOutput
}

type TXOutput struct {
    Value        int
    ScriptPubKey string // don't have addresses implemented, we'll avoid the whole scripting related logic for now.it will store an arbitary string
}

type TXInput struct {
    Txid      []byte   // stores the ID of such transaction,
    Vout      int      // stores an index of an output in the transaction. 
    ScriptSig string   //ScriptSig is a script which provides data to be used in an output’s ScriptPubKey. If the data is correct, the output can be unlocked, and its value can be used to generate new outputs
}


// A coinbase transaction has only one input.
func NewCoinbaseTX(to, data string) *Transaction {
    if data == "" {
            data = fmt.Sprintf("Reward to '%s'", to)
    }
        
    txin := TXInput{[]byte{}, -1, data}
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


