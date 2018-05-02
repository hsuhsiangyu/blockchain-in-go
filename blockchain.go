package main

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
    tip []byte
    db  *bolt.DB
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(data string) {
    var lastHash []byte

    err := bc.db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))
            lastHash = b.Get([]byte("l"))
                
            return nil
    })
    
    if err != nil {
            log.Panic(err)
    }
//After mining a new block, we save its serialized representation into the DB and update the l key, 
//which now stores the new block’s hash.
    newBlock := NewBlock(data, lastHash)
    
    err = bc.db.Update(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))
            err := b.Put(newBlock.Hash, newBlock.Serialize())
            if err != nil {
                    log.Panic(err)
            }
                            
            err = b.Put([]byte("l"), newBlock.Hash)
            if err != nil {
                    log.Panic(err)
            }
                                                    
            bc.tip = newBlock.Hash
                                                        
            return nil
    })
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
    var tip []byte
        db, err := bolt.Open(dbFile, 0600, nil)
        if err != nil {
            log.Panic(err)
        }
        err = db.Update(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))  // obtain the bucket storing our blocks
                if b == nil {
                    fmt.Println("No existing blockchain found. Creating a new one...")
                    genesis := NewGenesisBlock()
                    b, err := tx.CreateBucket([]byte(blocksBucket))  //create the bucket
                    if err != nil {
                            log.Panic(err)
                    }
                    err = b.Put(genesis.Hash, genesis.Serialize())
                    if err != nil {
                            log.Panic(err)
                    }

                    err = b.Put([]byte("l"), genesis.Hash)  //update the l key storing the last block hash of the chain.
                    if err != nil {
                            log.Panic(err)
                    }
                    tip = genesis.Hash
                } else {
                    tip = b.Get([]byte("l"))  
                }

                return nil
            })
            if err != nil {
                    log.Panic(err)
            }

            bc := Blockchain{tip, db} //only the tip of the chain is stored. Also, we store a DB connection, 

            return &bc
}













