package main

import "bytes"


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
