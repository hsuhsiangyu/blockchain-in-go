package main

import "bytes"


// TXInput represents a transaction input
type TXInput struct {
    Txid      []byte   // stores the ID of such transaction,
    Vout      int      // stores an index of an output in the transaction. 
    Signature []byte 
    PubKey    []byte
}

// checks that an input uses a specific key to unlock an output
// UseKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
    lockingHash := HashPubKey(in.PubKey)

    return bytes.Compare(lockingHash, pubKeyHash) == 0
}
