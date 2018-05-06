package main

import (
        "bytes"
        "errors"
        "fmt"
        "encoding/hex"
        "log"
        "crypto/ecdsa"
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

// CreateBlockchain createss a new Blockchain DB
func CreateBlockchain(address string) *Blockchain {
    if dbExists() {
            fmt.Println("Blockchain already exists.")
            os.Exit(1)
    }
    var tip []byte
    cbtx := NewCoinbaseTX(address, genesisCoinbaseData)  // takes an address which will receive the reward for mining the genesis block.
    genesis := NewGenesisBlock(cbtx)
    db, err := bolt.Open(dbFile, 0600, nil)
    if err != nil {
        log.Panic(err)
    }
    err = db.Update(func(tx *bolt.Tx) error {  // open a read-write transaction
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

// NewBlockchain creates43 a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
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


//finds a transaction by ID (this requires iterating over all the blocks in the blockchain)
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
    bci := bc.Iterator()

    for {
            block := bci.Next()
            for _, tx := range block.Transactions {
                    if bytes.Compare(tx.ID, ID) == 0 {
                            return *tx, nil
                    }
            }
            if len(block.PrevBlockHash) == 0 {
                    break
            }
    }
    return Transaction{}, errors.New("Transaction is not found")
}


// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
    UTXO := make(map[string]TXOutputs)
    spentTXOs := make(map[string][]int)
    bci := bc.Iterator()
    for {
        block := bci.Next()
    
        for _, tx := range block.Transactions {
            txID := hex.EncodeToString(tx.ID)
                    
            Outputs:
                    for outIdx, out := range tx.Vout {
                    // Was the output spent?
                        if spentTXOs[txID] != nil {
                            for _, spentOutIdx := range spentTXOs[txID] {
                                    if spentOutIdx == outIdx {
                                            continue Outputs
                                    }
                            }
                        }
                        outs := UTXO[txID]
                        outs.Outputs = append(outs.Outputs, out)
                        UTXO[txID] = outs
                    }
                    if tx.IsCoinbase() == false {
                            for _, in := range tx.Vin {
                                    inTxID := hex.EncodeToString(in.Txid)
                                    spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                            }
                    }
        }
        if len(block.PrevBlockHash) == 0 {
                break
        }
    }
    return UTXO
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
    bci := &BlockchainIterator{bc.tip, bc.db}

    return bci
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) MineBlock(transaction []*Transaction) *Block {
    var lastHash []byte

    for _, tx := range transaction {
            if bc.VerifyTransaction(tx) != true {
                    log.Panic("ERROR: Invalid transaction")
            }
    }

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
    if err != nil {
            log.Panic(err)
    }

    return newBlock
}

// SignTransaction signs inputs of a Transaction
// SignTransaction takes a transaction, finds transactions it references, and signs it;
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
            prevTX, err := bc.FindTransaction(vin.Txid)
                if err != nil {
                        log.Panic(err)
                }
                prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }
                                
    tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
    if tx.IsCoinbase() {  // another bug fix
        return true
    }
    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
            prevTX, err := bc.FindTransaction(vin.Txid)
            if err != nil {
                    log.Panic(err)
            }
            prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }
                                
    return tx.Verify(prevTXs)
}

func dbExists() bool {
    if _, err := os.Stat(dbFile); os.IsNotExist(err) {
            return false
    }
        
    return true
}

// instead, we use UTXO set
// FindUnspentTransactions returns a list of transactions containing unspent outputs
/*func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
    var unspentTXs []Transaction
    spentTXOs := make(map[string][]int)  //该交易里TxInput的来源交易的Txid作为key, 来源交易中对应的Output的index作为value
    bci := bc.Iterator()

    for {
            block := bci.Next()  // check every block in a blockchain, from last block to genesis block
                for _, tx := range block.Transactions {  // check every tx in the block
                        txID := hex.EncodeToString(tx.ID)  // get the ID of the tx
                Outputs:
                        for outIdx, out := range tx.Vout {  // check every Vout in the tx
// Was the output spent? skip those that were referenced in inputs (their values were moved to other outputs, thus we cannot count them)
                            if spentTXOs[txID] != nil {   // 如果该交易有 Output 被使用了 
                                    for _, spentOut := range spentTXOs[txID] {
                                            if spentOut == outIdx {      // 用该out的输出index 和切片中的元素一一比较
                                                    continue Outputs    // 存在的话，就忽略该 out， 不是忽略这笔交易!
                                            }
                                    }
                            }
                            if out.IsLockedWithKey(pubKeyHash) {  
                                unspentTXs = append(unspentTXs, *tx)  // 只要有没使用过的并能解锁的Output都加进来
                            }
                        }
                        if tx.IsCoinbase() == false {//gather all inputs that could unlock outputs locked with the provided address,
                                                // this doesn’t apply to coinbase transactions, since they don’t unlock outputs
                                for _, in := range tx.Vin {  // check every input in Vin
                                        if in.UsesKey(pubKeyHash) {
                                                inTxID := hex.EncodeToString(in.Txid)  // in.Txid 指来源交易的Txid，如tx2中input0的Txid是tx0的交易ID
                                                spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) //看图，spentTXOs[tx0] 中有Output 0 的index
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
*/





