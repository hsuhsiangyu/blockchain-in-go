package main

import (
    "encoding/hex"
    "log"

    "github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
    Blockchain *Blockchain
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
    db := u.Blockchain.db
    counter := 0

    err := db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(utxoBucket))
            c := b.Cursor()
            
            for k, _ := c.First(); k != nil; k, _ = c.Next() {
                    counter++
            }
                                
            return nil
    })
    if err != nil {
            log.Panic(err)
    }
                                            
    return counter
}

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
    db := u.Blockchain.db
    bucketName := []byte(utxoBucket)
    // it removes the bucket if it exists
    err := db.Update(func(tx *bolt.Tx) error {
            err := tx.DeleteBucket(bucketName)  
            if err != nil && err != bolt.ErrBucketNotFound {
                    log.Panic(err)
            }
            _, err = tx.CreateBucket(bucketName)
            if err != nil {
                    log.Panic(err)
            }

            return nil
    })
    if err != nil {
            log.Panic(err)
            }
    UTXO := u.Blockchain.FindUTXO()
    // it gets all unspent outputs from blockchain, and finally it saves the outputs to the bucket.
    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketName)
                    
        for txID, outs := range UTXO {
            key, err := hex.DecodeString(txID)
            if err != nil {
                    log.Panic(err)
            }
            err = b.Put(key, outs.Serialize())
            if err != nil {
                    log.Panic(err)
            }
        }
        return nil
    })
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
    unspentOutputs := make(map[string][]int)
    accumulated := 0
    db := u.Blockchain.db

    err := db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(utxoBucket))
            c := b.Cursor() // 要遍历键，我们将使用游标Cursor()
            
            for k, v := c.First(); k != nil; k, v = c.Next() { //First()  移动到第一个健.
                txID := hex.EncodeToString(k)
                outs := DeserializeOutputs(v)
                                    
                for outIdx, out := range outs.Outputs {
                    if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
                        accumulated += out.Value
                        unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
                    }
                }
            }
            return nil
    })
    if err != nil {
            log.Panic(err)
    }

    return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
    var UTXOs []TXOutput
    db := u.Blockchain.db

    err := db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(utxoBucket))
            c := b.Cursor()
            
            for k, v := c.First(); k != nil; k, v = c.Next() {
                outs := DeserializeOutputs(v)
                            
                for _, out := range outs.Outputs {
                    if out.IsLockedWithKey(pubKeyHash) {
                        UTXOs = append(UTXOs, out)
                    }
                }
            }
            return nil
    })
    if err != nil {
        log.Panic(err)
    }
    return UTXOs
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
    db := u.Blockchain.db

    err := db.Update(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(utxoBucket))
        for _, tx := range block.Transactions {
            if tx.IsCoinbase() == false {
                for _, vin := range tx.Vin {
                        updatedOuts := TXOutputs{}
                        outsBytes := b.Get(vin.Txid)  // Txid means the previous transaction ID
                        outs := DeserializeOutputs(outsBytes) // previous transaction output slice
// If a transaction which outputs were removed, contains no more outputs, it’s removed as well. ???????????                                                                              
                        for outIdx, out := range outs.Outputs {
                            if outIdx != vin.Vout {  
                                updatedOuts.Outputs = append(updatedOuts.Outputs, out)
                            }
                        }
                        if len(updatedOuts.Outputs) == 0 {
                            err := b.Delete(vin.Txid)
                            if err != nil {
                                    log.Panic(err)
                            }
                        } else {
                            err := b.Put(vin.Txid, updatedOuts.Serialize())
                            if err != nil {
                                    log.Panic(err)
                            }
                        }
                }
            }
            newOutputs := TXOutputs{}
            for _, out := range tx.Vout {  // add all the output of this transaction
                    newOutputs.Outputs = append(newOutputs.Outputs, out)    
            }

            err := b.Put(tx.ID, newOutputs.Serialize())
            if err != nil {
                    log.Panic(err)
            }
        }
        return nil
    })
    if err != nil {
            log.Panic(err)
    }
}




