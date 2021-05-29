package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	blockchain *Blockchain
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateblockchainwithgenesis -data -- 交易数据.")
	fmt.Println("\taddblock -data DATA -- 交易数据.")
	fmt.Println("\tprintchain -- 输出区块信息.")
}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {
	if DBExists() == false {
		fmt.Println("数据不存在.......")
		os.Exit(1)
	}

	blockchain := BlockchainObject()
	defer blockchain.DB.Close()

	blockchain.AddBlockToBlockchain(data)
}

func (cli *CLI) printchain() {
	if DBExists() == false {
		fmt.Println("数据不存在.......")
		os.Exit(1)
	}
	blockchain := BlockchainObject()
	defer blockchain.DB.Close()

	blockchain.Print()

}

func (cli *CLI) createGenesisBlockchain(data string) {
	CreateBlockchainWithGenesisBlock(data)
}

func (cli *CLI) Run() {
	isValidArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	flagCreateBlockchainWithData := createBlockchainCmd.String("data", "Genesis block data......", "创世区块交易数据......")
	flagAddBlockData := addBlockCmd.String("data", "A 给 B 10BTC", "交易数据......")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *flagCreateBlockchainWithData == "" {
			fmt.Println("交易数据不能为空......")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockchainWithData)
	}

	if addBlockCmd.Parsed() {
		if *flagAddBlockData == "" {
			printUsage()
			os.Exit(1)
		}
		cli.addBlock(*flagAddBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printchain()
	}
}
