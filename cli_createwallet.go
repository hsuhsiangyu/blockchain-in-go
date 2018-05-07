package main

import (
        "fmt"
)

func (cli *CLI) createWallet(nodeID string){
    wallets, _ := NewWallets(nodeID)         // fix the bug
    address := wallets.CreateWallet()
    wallets.SaveToFile(nodeID)        

    fmt.Printf("Your new address: %s\n", address)
}

