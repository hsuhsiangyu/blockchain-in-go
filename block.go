package main

import (
        "time"
    )

// Block keeps block headers
type Block struct {
    Timestamp     int64
    Data          []byte
    PrevBlockHash []byte
    Hash          []byte
    Nonce         int 
}


// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte) *Block {
    block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
    pow := NewProofOfWork(block)
    nonce, hash := pow.Run()

    block.Hash = hash[:]
    block.Nonce = nonce

    return block
}
// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock() *Block {
    return NewBlock("Genesis Block", []byte{})
}


func (b *Block) Serialize() []byte {
    var result bytes.Buffer         // declare a buffer that will store serialized data
    encoder := gob.NewEncoder(&result)  //initialize a gob encoder 

    err := encoder.Encode(b)        // encode the block

    return result.Bytes()           // the result is returned as a byte array
}


func DeserializeBlock(d []byte) *Block {
    var block Block

    decoder := gob.NewDecoder(bytes.NewReader(d))
    err := decoder.Decode(&block)

    return &block
}


