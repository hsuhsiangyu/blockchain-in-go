package main

import (
        "fmt"
)

func (cli *CLI) createWallet(){
    wallets, _ := NewWallets()         // fix the bug
    address := wallets.CreateWallet()
    wallets.SaveToFile()        

    fmt.Printf("Your new address: %s\n", address)
}

