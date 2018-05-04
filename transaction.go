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
    Value        int
    ScriptPubKey string // don't have addresses implemented, we'll avoid the whole scripting related logic for now.it will store an arbitary string
}

type TXInput struct {
    Txid      []byte   // stores the ID of such transaction,
    Vout      int      // stores an index of an output in the transaction. 
    ScriptSig string   //ScriptSig is a script which provides data to be used in an output’s ScriptPubKey. If the data is correct, the output can be unlocked, and its value can be used to generate new outputs
}

//just compare the script fields with unlockingData. These pieces will be improved in future
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
    return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
    return out.ScriptPubKey == unlockingData
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

