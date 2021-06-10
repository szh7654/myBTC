package BLC

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type CLI struct{}

func (cli *CLI) Run() {
	isValidArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("没有设置NODE_ID")
		os.Exit(1)
	}

	fmt.Println("当前节点是:", nodeID)

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getAddresslistsCmd := flag.NewFlagSet("getaddresslists", flag.ExitOnError)
	createblockchainCmd := flag.NewFlagSet("create", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	resetCmd := flag.NewFlagSet("reset", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	setCoinbaseCmd := flag.NewFlagSet("coinbase", flag.ExitOnError)

	flagFrom := sendCmd.String("from", "", "转账源地址")
	flagTo := sendCmd.String("to", "", "转账目的地址")
	flagAmount := sendCmd.String("amount", "", "转账金额")
	flagMine := sendCmd.String("mine", "", "本地挖矿")

	flagCoinbase := createblockchainCmd.String("address", "", "创世区块数据的地址")
	flagGetbalanceWithAddress := getBalanceCmd.String("address", "", "要查询余额的账户.......")
	flagStartNodeWithMiner := startNodeCmd.String("miner", "", "挖矿奖励的地址")
	flagSetCoinbaseWithAddress := setCoinbaseCmd.String("address", "", "挖矿奖励地址")

	switch os.Args[1] {
	case "send":
		if err := sendCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "create":
		if err := createblockchainCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "print":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "getbalance":
		if err := getBalanceCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "createwallet":
		if err := createWalletCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "getaddresslists":
		if err := getAddresslistsCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "reset":
		if err := resetCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "startnode":
		if err := startNodeCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "coinbase":
		if err := setCoinbaseCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	if sendCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}
		from := JSONToArray(*flagFrom)
		to := JSONToArray(*flagTo)
		amount := JSONToArray(*flagAmount)
		mine := true
		if *flagMine == "false" || *flagMine == "f" {
			mine = false
		}
		cli.send(from, to, amount, nodeID, mine)
	}

	if createblockchainCmd.Parsed() {
		if *flagCoinbase == "" {
			fmt.Println("地址不能为空....")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCoinbase, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printchain(nodeID)
	}

	if getBalanceCmd.Parsed() {
		if *flagGetbalanceWithAddress == "" {
			fmt.Println("地址不能为空....")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetbalanceWithAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet(nodeID)
	}

	if getAddresslistsCmd.Parsed() {
		cli.GetAddressList(nodeID)
	}

	if resetCmd.Parsed() {
		cli.Reset(nodeID)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(nodeID, *flagStartNodeWithMiner)
	}

	if setCoinbaseCmd.Parsed() {
		cli.setCoinbase(nodeID, *flagSetCoinbaseWithAddress)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreatewallet -- 创建钱包")
	fmt.Println("\tgetaddresslists -- 获取所有的钱包地址")
	fmt.Println("\tcreate -address --创世区块交易数据.")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -mine true/false  --转账交易")
	fmt.Println("\tprint --输出区块信息.")
	fmt.Println("\tgetbalance -address --获取address的余额.")
	fmt.Println("\treset --重置UTXOSet.")
	fmt.Println("\tstartnode --启动节点")
	fmt.Println("\tcoinbase -address --设置挖矿奖励地址")
}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) setCoinbase(nodeID string, coinbase string) {
	//将coinbase写入文件
	fileName := fmt.Sprintf("coinbase_%s", nodeID)
	data := []byte(coinbase)
	if ioutil.WriteFile(fileName, data, 0644) == nil {
		fmt.Println("写入文件成功:", coinbase)
	}
}

func CoinbaseAddress(nodeID string) string {
	fileName := fmt.Sprintf("coinbase_%s", nodeID)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Panic(err)
	}

	result := strings.Replace(string(b), "\n", "", 1)
	return result
}

func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	CreateBlockchainWithGenesisBlock(address, nodeID)

	blockchain := BlockchainObject(nodeID)
	//defer blockchain.DB.Close()

	if blockchain == nil {
		os.Exit(1)
	}

	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}

func (cli *CLI) CreateWallet(nodeID string) {
	wallets := NewWallets(nodeID)
	wallets.CreateWallet(nodeID)
	fmt.Println("钱包：", wallets.WalletMap)
}

func (cli *CLI) GetAddressList(nodeID string) {
	fmt.Println("打印所有的钱包地址。。")
	wallets := NewWallets(nodeID)
	for address := range wallets.WalletMap {
		fmt.Println("address: ", address)
	}
}

func (cli *CLI) getBalance(address string, nodeID string) {
	blockchain := BlockchainObject(nodeID)
	//defer blockchain.DB.Close()

	if blockchain == nil {
		os.Exit(1)
	}

	//txs 传nil值，查询时没有新的交易产生
	//total := blockchain.GetBalance(address, []*Transaction{})
	utxoSet := &UTXOSet{blockchain}
	total := utxoSet.GetBalance(address)
	fmt.Printf("%s的余额：%d\n", address, total)
}
func (cli *CLI) printchain(nodeID string) {
	blockchain := BlockchainObject(nodeID)
	//defer blockchain.DB.Close()

	if blockchain == nil {
		os.Exit(1)
	}

	blockchain.Printchain()
}

func (cli *CLI) Reset(nodeID string) {
	blockchain := BlockchainObject(nodeID)
	//defer blockchain.DB.Close()

	if blockchain == nil {
		os.Exit(1)
	}

	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mine bool) {
	/*
		address:  1Rs9zcPDqosXucdJjGP4wjGrtA1SmYpwGnQBMCprE2TdvhUyhk	c
		address:  1YfMAGkzTU3P19DobiAiggGzzcymvJyePughP37efhVgCV4W8e	b
		address:  1Z4DNkwSLgQR8yhtTnZSyobdenW3FfjMtkAnHJdM9ZAVenYDsU  	a
	*/
	//go run main.go send -from '["1XtLrwjcCnaBfE3Hwuypzchnsz7PKQLxnDyfba67cBkmXG1XYa"]' -to '["1RAbXZJVTYvPfdrXRd274vjrBE6XWxeyMLSxNYZixsuT7Uetrc"]' -amount '["1"]'
	//go run main.go send -from '["1XtLrwjcCnaBfE3Hwuypzchnsz7PKQLxnDyfba67cBkmXG1XYa","1XtLrwjcCnaBfE3Hwuypzchnsz7PKQLxnDyfba67cBkmXG1XYa"]' -to '["1RAbXZJVTYvPfdrXRd274vjrBE6XWxeyMLSxNYZixsuT7Uetrc","1RAbXZJVTYvPfdrXRd274vjrBE6XWxeyMLSxNYZixsuT7Uetrc"]' -amount '["2","1"]'
	//go run main.go send -from '["1YfMAGkzTU3P19DobiAiggGzzcymvJyePughP37efhVgCV4W8e","1Rs9zcPDqosXucdJjGP4wjGrtA1SmYpwGnQBMCprE2TdvhUyhk"]' -to '["1Rs9zcPDqosXucdJjGP4wjGrtA1SmYpwGnQBMCprE2TdvhUyhk","1Z4DNkwSLgQR8yhtTnZSyobdenW3FfjMtkAnHJdM9ZAVenYDsU"]' -amount '["3","1"]'
	//go run main.go send -from '["1Z4DNkwSLgQR8yhtTnZSyobdenW3FfjMtkAnHJdM9ZAVenYDsU"]' -to '["1Rs9zcPDqosXucdJjGP4wjGrtA1SmYpwGnQBMCprE2TdvhUyhk"]' -amount '["8"]'
	/*
		1/	a->b 4					a: 16 / b: 4 / c: 0
		2/	a->b 2  a->c 1			a: 23 / b: 6 / c: 1
		3/	b->c 3  c->a 1			a: 24 / b: 13 / c: 3
		4/  a->c 8					a: 26 / b: 13 / c: 11
	*/

	blockchain := BlockchainObject(nodeID)
	//defer blockchain.DB.Close()

	if blockchain == nil {
		os.Exit(1)
	}

	if mine {
		fmt.Println("--------------本地挖矿")
		txs := blockchain.hanldeTransations(from, to, amount, nodeID)
		newBlock := blockchain.MineNewBlock(txs)
		blockchain.SaveNewBlockToBlockchain(newBlock)
		utxoSet := &UTXOSet{blockchain}
		utxoSet.Update()
	} else {
		fmt.Println("--------------挖矿节点挖矿")
		//拼接nodeID到ip后
		nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
		txs := blockchain.hanldeTransations(from, to, amount, nodeID)
		//fmt.Println(nodeAddress)
		if nodeAddress != knowNodes[0] {
			//非主节点的交易先发送给主节点
			fmt.Println("sendTransactionToMainNode")
			sendTransactionToMainNode(knowNodes[0], txs)
		} else {
			//如果交易是在主节点, 直接发送给挖矿节点
			fmt.Println("sendTransactionToMiner")
			sendTransactionToMiner(knowNodes[1], txs)
		}
	}

}

func (cli *CLI) startNode(nodeID string, miner string) {
	//fmt.Println(miner)
	if miner == "" || IsValidAddress([]byte(miner)) {
		startServer(nodeID, miner)
	} else {
		fmt.Println("Miner地址无效")
		os.Exit(1)
	}
}
