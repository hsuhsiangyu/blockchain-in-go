package main

import (
        "time"
        "bytes"
        "encoding/gob"
        "log"
        "crypto/sha256"
    )

// Block keeps block headers
type Block struct {
    Timestamp     int64
    Transaction   []*Transaction
    PrevBlockHash []byte
    Hash          []byte
    Nonce         int 
}


// NewBlock creates and returns Block
func NewBlock(transaction []*Transaction, prevBlockHash []byte) *Block {
    block := &Block{time.Now().Unix(), transaction, prevBlockHash, []byte{}, 0} // []byte can be initiled by string
    pow := NewProofOfWork(block)  // obtain a pow struct which contains block pointer and target
    nonce, hash := pow.Run()

    block.Hash = hash[:]
    block.Nonce = nonce

    return block
}
// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
    return NewBlock([]*Transaction{coinbase}, []byte{})  // Genesis Block's preBlockHash must be []byte{}
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
    var txHashes [][]byte
    var txHash [32]byte

    for _, tx := range b.Transaction {
            txHashes = append(txHashes, tx.ID)
    }
    txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
        
    return txHash[:]
}


func (b *Block) Serialize() []byte {
    var result bytes.Buffer         // declare a buffer that will store serialized data
    encoder := gob.NewEncoder(&result)  //initialize a gob encoder 

    err := encoder.Encode(b)        // encode the block
    if err != nil {
            log.Panic(err)
    }    
    return result.Bytes()           // the result is returned as a byte array
}


func DeserializeBlock(d []byte) *Block {
    var block Block

    decoder := gob.NewDecoder(bytes.NewReader(d))
    err := decoder.Decode(&block)
    if err != nil {
            log.Panic(err)
    }
    return &block
}


