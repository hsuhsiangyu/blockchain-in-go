package main

import (
    "bytes"
    "encoding/gob"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var blocksInTransit = [][]byte{}
var knownNodes = []string{"localhost:3000"}  // hardcode the address of the central node: every node must know where to connect to initially
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type version struct {
    Version    int
    BestHeight int
    AddrFrom   string
}

type getblocks struct {
    AddrFrom string
}

type addr struct {
    AddrList []string
}

type inv struct {  //uses inv to show other nodes what blocks or transactions current node has. just their hashes
    AddrFrom string
    Type     string   //The Type field says whether these are blocks or transactions.
    Items    [][]byte
}

//getdata is a request for certain block or transaction, and it can contain only one block/transaction ID.
type getdata struct {
    AddrFrom string
    Type     string
    ID       []byte
}

type block struct {
    AddrFrom string
    Block    []byte
}

type tx struct {
    AddFrom     string
    Transaction []byte
}


func StartServer(nodeID, minerAddress string) {
    nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
    miningAddress = minerAddress
    ln, err := net.Listen(protocol, nodeAddress)
    defer ln.Close()

    bc := NewBlockchain(nodeID)

    if nodeAddress != knownNodes[0] { // if current node is not the central one, it must send version message to the central node to find out if its blockchain is outdated
        sendVersion(knownNodes[0], bc)
    }
        
    for {
        conn, err := ln.Accept()
        go handleConnection(conn, bc)
    }
}






func sendVersion(addr string, bc *Blockchain) {
    bestHeight := bc.GetBestHeight()
    payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})
//First 12 bytes specify command name (“version” in this case), and the latter bytes will contain gob-encoded message structure.
    request := append(commandToBytes("version"), payload...)

    sendData(addr, request)
}

func commandToBytes(command string) []byte {
    var bytes [commandLength]byte

    for i, c := range command {
        bytes[i] = byte(c)
    }
        
    return bytes[:]
}

func bytesToCommand(bytes []byte) string {
    var command []byte

    for _, b := range bytes {
            if b != 0x0 {
                command = append(command, b)
            }
    }
                        
    return fmt.Sprintf("%s", command)
}

func handleConnection(conn net.Conn, bc *Blockchain) {
    request, err := ioutil.ReadAll(conn) //When a node receives a command, it runs bytesToCommand to extract command name and processes command body with correct handler
    command := bytesToCommand(request[:commandLength])
    fmt.Printf("Received %s command\n", command)

    switch command {
        ...
        case "version":
            handleVersion(request, bc)
        default:
            fmt.Println("Unknown command!")
    }
                    
    conn.Close()
}

func handleVersion(request []byte, bc *Blockchain) {
    var buff bytes.Buffer
    var payload verzion

    buff.Write(request[commandLength:])
    dec := gob.NewDecoder(&buff) //decode the request and extract the payload
    err := dec.Decode(&payload)

    myBestHeight := bc.GetBestHeight()
    foreignerBestHeight := payload.BestHeight
// compares its BestHeight with the one from the message
    if myBestHeight < foreignerBestHeight {
        sendGetBlocks(payload.AddrFrom)
    } else if myBestHeight > foreignerBestHeight {
        sendVersion(payload.AddrFrom, bc)
    }
                
    if !nodeIsKnown(payload.AddrFrom) {
        knownNodes = append(knownNodes, payload.AddrFrom)
    }
}

// it requests a list of block hashes. This is done to reduce network load, because blocks can be downloaded from different nodes, and we don’t want to download dozens of gigabytes from one node.
func handleGetBlocks(request []byte, bc *Blockchain) {
    ...
    blocks := bc.GetBlockHashes()
    sendInv(payload.AddrFrom, "block", blocks)
}


func handleInv(request []byte, bc *Blockchain) {
    ...
    fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

    if payload.Type == "block" {
            blocksInTransit = payload.Items
//If blocks hashes are transferred, we want to save them in blocksInTransit variable to track downloaded blocks.        
            blockHash := payload.Items[0]
//we send getdata command to the sender of the inv message and update blocksInTransit
            sendGetData(payload.AddrFrom, "block", blockHash)
                
            newInTransit := [][]byte{}
            for _, b := range blocksInTransit {
                    if bytes.Compare(b, blockHash) != 0 {
                            newInTransit = append(newInTransit, b) // add the other blockhash in payload.Items, without b
                    }
            }
            blocksInTransit = newInTransit
    }
    if payload.Type == "tx" {
        txID := payload.Items[0] //we’ll never send inv with multiple hashes. That’s why only the first hash is taken
    
        if mempool[hex.EncodeToString(txID)].ID == nil {
                sendGetData(payload.AddrFrom, "tx", txID)
        }
    }
}

//we don’t check if we actually have this block or transaction. This is a flaw
func handleGetData(request []byte, bc *Blockchain) {
    ...
    if payload.Type == "block" {
        block, err := bc.GetBlock([]byte(payload.ID))
        
        sendBlock(payload.AddrFrom, &block)
    }
            
    if payload.Type == "tx" {
        txID := hex.EncodeToString(payload.ID)
        tx := mempool[txID]
                        
        sendTx(payload.AddrFrom, &tx)
    }
}
//TODO: Instead of trusting unconditionally, we should validate every incoming block before adding it to the blockchain.
//TODO: Instead of running UTXOSet.Reindex(), UTXOSet.Update(block) should be used, because if blockchain is big, it’ll take a lot of time to reindex the whole UTXO set.
func handleBlock(request []byte, bc *Blockchain) {
    ...

    blockData := payload.Block
    block := DeserializeBlock(blockData)

    fmt.Println("Recevied a new block!")
    bc.AddBlock(block)

    fmt.Printf("Added block %x\n", block.Hash)

    if len(blocksInTransit) > 0 {
            blockHash := blocksInTransit[0]
            sendGetData(payload.AddrFrom, "block", blockHash)
//If there’re more blocks to download, we request them from the same node we downloaded the previous block. 
            blocksInTransit = blocksInTransit[1:]
    } else {
            UTXOSet := UTXOSet{bc} //When we finally downloaded all the blocks, the UTXO set is reindexed.
            UTXOSet.Reindex()
    }
}

func handleTx(request []byte, bc *Blockchain) {
    ...
    txData := payload.Transaction
    tx := DeserializeTransaction(txData)
    mempool[hex.EncodeToString(tx.ID)] = tx //to put new transaction in the mempool 

    if nodeAddress == knownNodes[0] {  // Checks whether the current node is the central one
        for _, node := range knownNodes {
            if node != nodeAddress && node != payload.AddFrom {
                sendInv(node, "tx", [][]byte{tx.ID}) //he central node won’t mine blocks. Instead, it’ll forward the new transactions to other nodes
            }
        }
    }else {
        if len(mempool) >= 2 && len(miningAddress) > 0 { //When there are 2 or more transactions in the mempool of the current (miner) node, mining begins.
            MineTransactions:
                var txs []*Transaction
                            
                for id := range mempool {
                    tx := mempool[id]
                    if bc.VerifyTransaction(&tx) {
                        txs = append(txs, &tx)
                    }
                }
                if len(txs) == 0 {
                    fmt.Println("All transactions are invalid! Waiting for new ones...")
                    return
                }
                cbTx := NewCoinbaseTX(miningAddress, "")
                txs = append(txs, cbTx) //Verified transactions are being put into a block,as well as a coinbase transaction with the reward
                newBlock := bc.MineBlock(txs)
                UTXOSet := UTXOSet{bc}
                UTXOSet.Reindex()
                fmt.Println("New block is mined!")
// After a transaction is mined, it’s removed from the mempool.
                for _, tx := range txs {
                    txID := hex.EncodeToString(tx.ID)
                    delete(mempool, txID)
                }
//Every other nodes the current node is aware of, receive inv message with the new block’s hash. They can request the block after handling the message.
                for _, node := range knownNodes {
                        if node != nodeAddress {
                                sendInv(node, "block", [][]byte{newBlock.Hash})
                        }
                }

                if len(mempool) > 0 {
                        goto MineTransactions
                }
        }
    }
}

func extractCommand(request []byte) []byte {
    return request[:commandLength]
}

func requestBlocks() {
    for _, node := range knownNodes {
            sendGetBlocks(node)
    }
}
