package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
}

func (cli *CLI) createBlockchain(address string) {
	bc := NewBlockChain(address)
	UTXOSet := UTXOSet{bc}
	UTXOSet.ReIndex()
	bc.db.Close()
	fmt.Printf("Blockchain created.")
}

func (cli *CLI) Run() {

	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWallet := flag.NewFlagSet("createwallet", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	createblockchain := flag.NewFlagSet("createblockchain", flag.ExitOnError)

	address := createblockchain.String("address", "", "address for genesis coinbase transaction")

	getBalance := flag.NewFlagSet("getbalance", flag.ExitOnError)
	balanceAddress := getBalance.String("address", "", "address for balance")

	send := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := send.String("from", "", "address for from")
	sendTo := send.String("to", "", "address for to")
	sendAmount := send.String("amount", "", "amount to transfer")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "createwallet":
		err := createWallet.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "createblockchain":
		err := createblockchain.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "getbalance":
		err := getBalance.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "send":
		err := send.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createblockchain.Parsed() {
		if *address == "" {
			createblockchain.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*address)
	}
	if getBalance.Parsed() {
		if *balanceAddress == "" {
			getBalance.Usage()
			os.Exit(1)
		}
		cli.getBalance(*balanceAddress)
	}
	if send.Parsed() {
		if *sendTo == "" || *sendFrom == "" || *sendAmount == "" {
			send.Usage()
			os.Exit(1)
		}
		amnt, err := strconv.Atoi(*sendAmount)
		if err != nil {
			panic(err)
		}
		cli.send(*sendFrom, *sendTo, amnt)
	}
	if createWallet.Parsed() {
		cli.createWallet()
	}

}

func (cli *CLI) validateArgs() bool {
	return true
}
func (cli *CLI) printUsage() {
	fmt.Printf("addblock with -data\n")
	fmt.Printf("printchain")
	fmt.Printf("createchain -address ADDRESS")
}

func (cli *CLI) addBlock(data string) {

	// cli.bc.AddBlock(data)
	fmt.Println("Success")

}

func (cli *CLI) printChain() {

	bc := NewBlockChain("")
	defer bc.db.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		// fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		fmt.Printf("Transactions: \n")
		for _, tx := range block.Transactions {
			fmt.Printf("\tID:%x\n", tx.ID)
			for _, vout := range tx.Vout {
				fmt.Printf("\t>Value: %d\n", vout.Value)
			}
			fmt.Println()
		}
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

}

func (cli *CLI) getBalance(address string) {

	bc := NewBlockChain(address)

	defer bc.db.Close()

	balance := 0

	pubKeyHash := Base58Decode([]byte(address))

	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOSet := UTXOSet{bc}
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s' : %d\n", address, balance)

}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)

	cbTx := NewCoinbaseTx(from, "")

	newBlock := bc.MineBlock([]*Transaction{cbTx, tx})

	UTXOSet := UTXOSet{bc}
	UTXOSet.Update(newBlock)

	fmt.Println("Success")

}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your wallet address is: %s\n", address)
}
