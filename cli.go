package main

import (
    "flag"
    "fmt"
    "log"
    "os"
)

type CLI struct {}

func (cli *CLI) printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  printchain - print all the blocks of the blockchain")
    fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
    fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
    fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
    fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
    fmt.Println("  listaddresses - Lists all addresses from the wallet file")
    fmt.Println("  reindexutxo - Rebuilds the UTXO set")
}

func (cli *CLI) validateArgs() { 
    if len(os.Args) < 2 {   // must hava more than one parament
            cli.printUsage()
            os.Exit(1)
    }
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
    cli.validateArgs()
    
    createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
    reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError) 

    createBlockchainAddress := createBlockchainCmd.String("address", "", "he address to send genesis block reward to")
    getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
    sendFrom := sendCmd.String("from", "", "Source wallet address")
    sendTo := sendCmd.String("to", "", "Destination wallet address")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")

    //check the command provided by user and parse related flag subcommand.
    switch os.Args[1] {
    case "createblockchain":
            err := createBlockchainCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "printchain":
            err := printChainCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "getbalance":
            err := getBalanceCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "send" :
            err := sendCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "createwallet":
            err := createWalletCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "listaddresses":
            err := listAddressesCmd.Parse(os.Args[2:])
            if err != nil {
                    log.Panic(err)
            }
    case "reindexutxo":
            err := reindexUTXOCmd.Parse(os.Args[2:])
            if err != nil {
                        log.Panic(err)
            }
    default:
            cli.printUsage()
            os.Exit(1)
    }

    if createBlockchainCmd.Parsed() {
            if *createBlockchainAddress == "" {
                    createBlockchainCmd.Usage()
                    os.Exit(1)
            }
            cli.createBlockchain(*createBlockchainAddress)
    }
    if printChainCmd.Parsed() {
            cli.printChain()
    }
    if getBalanceCmd.Parsed() {
            if *getBalanceAddress == "" {
                    getBalanceCmd.Usage()
                    os.Exit(1)
            }
            cli.getBalance(*getBalanceAddress)
    }
    if sendCmd.Parsed() {
            if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
                    sendCmd.Usage()
                    os.Exit(1)
            }
            // here pay attenion 
            cli.send(*sendFrom, *sendTo, *sendAmount)
    }
    if createWalletCmd.Parsed() {
            cli.createWallet()
    }
    if listAddressesCmd.Parsed() {
            cli.listAddresses()
    }
    if reindexUTXOCmd.Parsed() {
            cli.reindexUTXO()
    }
}
