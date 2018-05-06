package main

import (
        "bytes"
        "log"
        "encoding/gob"
)

type TXOutput struct {
    Value       int
    PubKeyHash  []byte //There are no real inputs in coinbase transactions, so signing is not necessary. The output of the coinbase transaction contains a hashed public key 
}

// Lock simply locks an output. When we send coins to someone, we know only their address, thus the function takes an address as the only argument.
func (out *TXOutput) Lock(address []byte) {
    pubKeyHash := Base58Decode(address)
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    out.PubKeyHash = pubKeyHash
}

// checks if provided public key hash was used to lock the output. 
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
    return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
    txo := &TXOutput{value, nil}
    txo.Lock([]byte(address))

    return txo
}

// TXOutputs collects TXOutput
type TXOutputs struct {
    Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
    var buff bytes.Buffer

    enc := gob.NewEncoder(&buff)
    err := enc.Encode(outs)
    if err != nil {
            log.Panic(err)
    }
        
    return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
    var outputs TXOutputs

    dec := gob.NewDecoder(bytes.NewReader(data))
    err := dec.Decode(&outputs)
    if err != nil {
            log.Panic(err)
    }
        
    return outputs
}

