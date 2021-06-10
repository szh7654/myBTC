package BLC

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type BlockData struct {
	from  string
	block []byte
}

type GetBlocks struct {
	from string
}

type GetData struct {
	from string
	Type string // Block or Transaction
	hash []byte //block或者Tx的hash
}

type Version struct {
	version int64
	height  int64
	from    string
}

type Inv struct {
	from  string
	Type  string   // Block or Transaction
	items [][]byte // Hash data
}

var knowNodes = []string{"localhost:3000", "localhost:3002", "localhost:3001"} //主节点地址/挖矿节点/普通节点

var nodeAddress string //当前节点地址

var blockArray [][]byte //记录尚未同步的区块的hash

var coinbaseAddress string //挖矿奖励分配地址

// Start a full node
func startServer(nodeID string, minerAddress string) {
	coinbaseAddress = minerAddress
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	listener, err := net.Listen("tcp", nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	bc := BlockchainObject(nodeID)

	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("New connection: ", conn.RemoteAddr())
		go handleConnection(conn, bc)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:COMMAND_LENGTH])

	fmt.Printf("New command recieved：%s\n", command)

	switch command {
	case COMMAND_VERSION:
		handleVersion(request, bc)
	case COMMAND_GETBLOCKS:
		handleGetBlocksHash(request, bc)
	case COMMAND_INV:
		handleInv(request, bc)
	case COMMAND_GETDATA:
		handleGetData(request, bc)
	case COMMAND_BLOCKDATA:
		handleGetBlockData(request, bc)
	case COMMAND_TXS:
		handleTransactions(request, bc)
	case COMMAND_REQUIREMINE:
		handleRequireMine(request, bc)
	case COMMAND_VERIFYBLOCK:
		handleVerifyBlock(request, bc)
	default:
		fmt.Println("无法识别....")
	}
	defer conn.Close()
}

// 本地高度>对方高度 -> 向对方发送本地的区块链消息
// 对方高度>本地高度 -> 向对方请求对方的区块链信息
func handleVersion(request []byte, bc *Blockchain) {
	fmt.Println("handle version")
	commandBytes := request[COMMAND_LENGTH:]
	var version Version
	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))
	err := decoder.Decode(&version)
	if err != nil {
		log.Panic(err)
	}

	height := bc.GetHeight()
	anotherHeight := version.height
	if height > anotherHeight {
		sendVersion(version.from, bc)
	} else if anotherHeight > height {
		sendGetBlocksHash(version.from)
	}
}

func handleGetBlocksHash(request []byte, bc *Blockchain) {
	fmt.Println(handleGetBlocksHash)
	commandBytes := request[COMMAND_LENGTH:]
	var getblocks GetBlocks

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))
	err := decoder.Decode(&getblocks)
	if err != nil {
		log.Panic(err)
	}
	// get hashes of all Blocks
	blocksHashes := bc.getAllBlocksHash()
	sendHash(getblocks.from, BLOCK_TYPE, blocksHashes)
}

/*
	处理Inv命令
	1. block type :  如果本地区块

*/
func handleInv(request []byte, bc *Blockchain) {
	commandBytes := request[COMMAND_LENGTH:]

	var inv Inv

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&inv)
	if err != nil {
		log.Panic(err)
	}

	if inv.Type == BLOCK_TYPE {
		//获取hashes中第一个hash,请求对方返回此hash对应的block
		hash := inv.items[0]
		sendGetData(inv.from, BLOCK_TYPE, hash)

		//保存items剩余未请求的hashes到变量blockArray(handleBlockData 方法会用到)
		if len(inv.items) > 0 {
			blockArray = inv.items[1:]
		}

	} else if inv.Type == TX_TYPE {

	}
}

func handleGetData(request []byte, bc *Blockchain) {
	//1.从request中获取版本的数据：[]byte
	commandBytes := request[COMMAND_LENGTH:]

	//2.反序列化--->version
	var getData GetData

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&getData)
	if err != nil {
		log.Panic(err)
	}

	if getData.Type == BLOCK_TYPE {
		block := bc.GetBlockByHash(getData.hash)
		sendBlock(getData.from, block)
	} else if getData.Type == TX_TYPE {

	}
}

func handleGetBlockData(request []byte, bc *Blockchain) {
	//1.从request中获取版本的数据：[]byte
	commandBytes := request[COMMAND_LENGTH:]

	//2.反序列化--->version
	var getBlockData BlockData

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&getBlockData)
	if err != nil {
		log.Panic(err)
	}

	blockBytes := getBlockData.block
	//block := DeserializeBlock(blockBytes)
	var block Block
	gobDecode(blockBytes, &block)
	//fmt.Println(&block)
	bc.AddBlockToChain(&block)

	if len(blockArray) == 0 {
		utxoSet := UTXOSet{bc}
		utxoSet.ResetUTXOSet()

	}

	if len(blockArray) > 0 {
		hash := blockArray[0]
		sendGetData(getBlockData.from, BLOCK_TYPE, hash)
		blockArray = blockArray[1:]
	}

}

/*
	主节点处理接收到的交易
*/
func handleTransactions(request []byte, bc *Blockchain) {
	//1.从request中获取版本的数据：[]byte
	commandBytes := request[COMMAND_LENGTH:]

	//2.反序列化--->version
	var txs []*Transaction

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&txs)
	if err != nil {
		log.Panic(err)
	}

	//发送到挖矿节点
	sendTransactionToMiner(knowNodes[1], txs)

	//for _, tx := range txs {
	//	//fmt.Println("处理获取到的txs")
	//	//fmt.Println(tx)
	//}
}

func handleRequireMine(request []byte, bc *Blockchain) {
	//1.从request中获取版本的数据：[]byte
	commandBytes := request[COMMAND_LENGTH:]
	//fmt.Println("反序列化得到的txbytes：")
	//fmt.Printf("%x",commandBytes)
	//fmt.Println("-----")

	//2.反序列化--->version
	var txs []*Transaction

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&txs)
	if err != nil {
		log.Panic(err)
	}

	//fmt.Printf("%x",gobEncode(txs))

	nodeID := os.Getenv("NODE_ID")
	txp := NewTXPool(nodeID)
	//将txs保存到交易池
	txp.Txs = append(txp.Txs, txs...)
	//for _, tx := range txp.Transactions {
	//	fmt.Println(tx)
	//}
	txp.saveFile(nodeID)

	const packageNum = 1

	//2. 判断交易池是否有足够的交易
	if len(txp.Txs) > 0 {
		//开始挖矿
		fmt.Println("开始挖矿")

		blockchain := BlockchainObject(nodeID)

		//取出要打包的交易
		//packageTx := txp.Transactions[:packageNum]
		newBlock := blockchain.MineNewBlock(txs)
		//fmt.Println(newBlock)
		txp.Txs = txp.Txs[packageNum:]
		txp.saveFile(nodeID)
		//发送newBlock 给主节点验证工作量证明
		sendNewBlockToMain(knowNodes[0], newBlock)
	}
}

func handleVerifyBlock(request []byte, blockchain *Blockchain) {
	//1.从request中获取版本的数据：[]byte
	commandBytes := request[COMMAND_LENGTH:]

	//2.反序列化--->version
	var block *Block

	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	pow := PoWFactory(block)
	if pow.IsValid() {
		blockchain.SaveNewBlockToBlockchain(block)
		utxoSet := &UTXOSet{blockchain}
		utxoSet.Update()

		//这里直接调起一次version命令  更新挖矿节点的区块
		sendVersion(knowNodes[1], blockchain)
	}

}

/*
	所有消息都是通过这个方法来发送到其他节点
*/
func sendData(to string, data []byte) {
	//fmt.Println("向",to,"发送",data)
	conn, err := net.Dial("tcp", to)
	if err != nil {
		log.Panic(err)
	}

	defer conn.Close()

	//发送数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

// Send the block height in current node
func sendVersion(to string, bc *Blockchain) {
	height := bc.GetHeight()
	version := &Version{NODE_VERSION, height, nodeAddress}
	sendCommandData(COMMAND_VERSION, version, to)
}

// Send message to get all blocks' hash
func sendGetBlocksHash(to string) {
	getBlocks := GetBlocks{nodeAddress}
	sendCommandData(COMMAND_GETBLOCKS, getBlocks, to)
}

// send all blocks' hash
func sendHash(to string, kind string, data [][]byte) {
	inv := Inv{nodeAddress, kind, data}
	sendCommandData(COMMAND_INV, inv, to)
}

/*
	发送请求对方根据hash返回对应的block的消息
*/
func sendGetData(to string, kind string, hash []byte) {
	//1.创建对象
	getData := GetData{nodeAddress, kind, hash}

	sendCommandData(COMMAND_GETDATA, getData, to)
}

/*
	发送block对象给对方
*/
func sendBlock(to string, block *Block) {
	//1.创建对象
	blockData := BlockData{nodeAddress, gobEncode(block)}

	sendCommandData(COMMAND_BLOCKDATA, blockData, to)
}

/*
	发送交易信息到主节点
*/
func sendTransactionToMainNode(to string, txs []*Transaction) {
	sendCommandData(COMMAND_TXS, txs, to)
}

func sendTransactionToMiner(to string, txs []*Transaction) {
	sendCommandData(COMMAND_REQUIREMINE, txs, to)
}

func sendNewBlockToMain(to string, block *Block) {
	sendCommandData(COMMAND_VERIFYBLOCK, block, to)
}

func sendCommandData(command string, data interface{}, to string) {
	//2.对象序列化为[]byte
	payload := gobEncode(data)
	//3.拼接命令和对象序列化
	request := append(commandToBytes(command), payload...)
	//4.发送消息
	sendData(to, request)
}
