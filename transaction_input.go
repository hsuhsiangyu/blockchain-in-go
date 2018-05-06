package main

import "bytes"


// TXInput represents a transaction input
type TXInput struct {
    Txid      []byte   // stores the ID of last transaction 来源交易的ID ,
    Vout      int      // stores an index of an output in the last transaction 即来源交易. 
    Signature []byte   // When one sends coins, a transaction is created. Inputs of the transaction will reference outputs from previous transaction(s). Every input will store a public key (not hashed) and a signature of the whole transaction.
    PubKey    []byte
}

// checks that an input uses a specific key to unlock an output
// UseKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
    lockingHash := HashPubKey(in.PubKey)

    return bytes.Compare(lockingHash, pubKeyHash) == 0
}
