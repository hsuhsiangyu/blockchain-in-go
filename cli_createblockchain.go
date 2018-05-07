package main

import (
        "fmt"
        "log"
)

func (cli *CLI) createBlockchain(address string, nodeID string) {
    if !ValidateAddress(address) {
            log.Panic("ERROR: Address is not valid")
    }
    bc := CreateBlockchain(address, nodeID)
    defer bc.db.Close()  // fix the bug, database not open
    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Println("Done!")
}


