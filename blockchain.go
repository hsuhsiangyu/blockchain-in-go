package main

import (
        "fmt"
        "encoding/hex"
        "log"
        "os"
        "github.com/boltdb/bolt"
    )

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"


// Blockchain keeps a sequence of Blocks
type Blockchain struct {
    tip []byte
    db  *bolt.DB
}

type BlockchainIterator struct {
    currentHash []byte
    db          *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
    bci := &BlockchainIterator{bc.tip, bc.db}

    return bci
}

func dbExists() bool {
    if _, err := os.Stat(dbFile); os.IsNotExist(err) {
            return false
    }
        
    return true
}


// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
    var block *Block

    err := i.db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))
            encodedBlock := b.Get(i.currentHash)
            block = DeserializeBlock(encodedBlock)
            return nil
    })
    if err != nil {
            log.Panic(err)
    }
    i.currentHash = block.PrevBlockHash
    return block
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) MineBlock(transaction []*Transaction) {
    var lastHash []byte

    err := bc.db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))  //obtain the bucket storing our blocks
            lastHash = b.Get([]byte("l"))
            return nil
    })
    if err != nil {
            log.Panic(err)
    }
//After mining a new block, we save its serialized representation into the DB and update the l key, 
//which now stores the new block’s hash.
    newBlock := NewBlock(transaction, lastHash)
    
    err = bc.db.Update(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte(blocksBucket))
            err := b.Put(newBlock.Hash, newBlock.Serialize()) // first, store newBlock.Hash as key, newBlock as value
            if err != nil {
                    log.Panic(err)
            }
            err = b.Put([]byte("l"), newBlock.Hash)  // second, store l as key, newBlock.Hash as value
            if err != nil {
                    log.Panic(err)
            }
            bc.tip = newBlock.Hash
            return nil
    })
}
// have some confuse
// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
    var unspentTXs []Transaction
    spentTXOs := make(map[string][]int)
    bci := bc.Iterator()

    for {
            block := bci.Next()  // check every block in a blockchain
                for _, tx := range block.Transactions {
                        txID := hex.EncodeToString(tx.ID)
                Outputs:
                        for outIdx, out := range tx.Vout {
                            // Was the output spent? skip those that were referenced in inputs (their values were moved to other outputs, thus we cannot count them)
                            if spentTXOs[txID] != nil { 
                                    for _, spentOut := range spentTXOs[txID] {
                                            if spentOut == outIdx {      // don't figure out ??????????????
                                                    continue Outputs
                                            }
                                    }
                            }
                            if out.CanBeUnlockedWith(address) {  
                                unspentTXs = append(unspentTXs, *tx)
                            }
                        }
                        if tx.IsCoinbase() == false {         //gather all inputs that could unlock outputs locked with the provided address,
                                                              // this doesn’t apply to coinbase transactions, since they don’t unlock outputs
                                for _, in := range tx.Vin {
                                        if in.CanUnlockOutputWith(address) {
                                                inTxID := hex.EncodeToString(in.Txid)
                                                spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                                        }      
                                }
                        }
            }

            if len(block.PrevBlockHash) == 0 {
                    break
            }
    }
    return unspentTXs
}


// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
    var UTXOs []TXOutput
    unspentTransactions := bc.FindUnspentTransactions(address)
      
    for _, tx := range unspentTransactions {
            for _, out := range tx.Vout {
                    if out.CanBeUnlockedWith(address) {
                            UTXOs = append(UTXOs, out)
                    }
            }
    }
    return UTXOs
}

// NewBlockchain creates43 a new Blockchain with genesis Block
func NewBlockchain(address string) *Blockchain {
    if dbExists() == false {
            fmt.Println("No existing blockchain found. Create one first.")
            os.Exit(1)
    }
    var tip []byte
        db, err := bolt.Open(dbFile, 0600, nil)
        if err != nil {
            log.Panic(err)
        }
        err = db.Update(func(tx *bolt.Tx) error {  // open a read-write transaction
                b := tx.Bucket([]byte(blocksBucket))  // obtain the bucket storing our blocks
                tip = b.Get([]byte("l"))  
                return nil
            })
            if err != nil {
                    log.Panic(err)
            }

            bc := Blockchain{tip, db} //only the tip of the chain is stored. Also, we store a DB connection, all block stored in DB

            return &bc
}


// CreateBlockchain createss a new Blockchain DB
func CreateBlockchain(address string) *Blockchain {
    if dbExists() {
            fmt.Println("Blockchain already exists.")
            os.Exit(1)
    }
    var tip []byte
        db, err := bolt.Open(dbFile, 0600, nil)
        if err != nil {
            log.Panic(err)
        }
        err = db.Update(func(tx *bolt.Tx) error {  // open a read-write transaction
                    cbtx := NewCoinbaseTX(address, genesisCoinbaseData)  // takes an address which will receive the reward for mining the genesis block.
                    genesis := NewGenesisBlock(cbtx)

                    b, err := tx.CreateBucket([]byte(blocksBucket))  //create the bucket
                    if err != nil {
                            log.Panic(err)
                    }
                    err = b.Put(genesis.Hash, genesis.Serialize())
                    if err != nil {
                            log.Panic(err)
                    }

                    err = b.Put([]byte("l"), genesis.Hash)  //update the l key storing the last block hash of the chain. ??
                    if err != nil {
                            log.Panic(err)
                    }
                    tip = genesis.Hash

                    return nil
            })
            if err != nil {
                    log.Panic(err)
            }

            bc := Blockchain{tip, db} //only the tip of the chain is stored. Also, we store a DB connection, all block stored in DB

            return &bc
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
    unspentOutputs := make(map[string][]int)
    unspentTXs := bc.FindUnspentTransactions(address)
    accumulated := 0

Work:
    for _, tx := range unspentTXs {  // iterates over all unspent transactions and accumulates their values
            txID := hex.EncodeToString(tx.ID)
        
            for outIdx, out := range tx.Vout {
                    if out.CanBeUnlockedWith(address) && accumulated < amount {
                            accumulated += out.Value
                            unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
                            if accumulated >= amount {
                                    break Work
                            }
                    }
            }
    }

    return accumulated, unspentOutputs
}








