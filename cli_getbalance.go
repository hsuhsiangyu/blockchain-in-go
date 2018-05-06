package main

import (
        "fmt"
        "log"
)

func (cli *CLI) getBalance(address string) {
    if !ValidateAddress(address){
            log.Panic("ERROR: Adderss is not valid")
    }
    bc := NewBlockchain()
    UXTOSet := UTXOSet{bc}
    defer bc.db.Close()

    balance := 0
    pubKeyHash := Base58Decode([]byte(address))
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    UTXOs := UXTOSet.FindUTXO(pubKeyHash)
    for _, out := range UTXOs {
            balance += out.Value
    }

    fmt.Printf("Balance of '%s': %d\n", address, balance)
}



