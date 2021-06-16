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
	commandBytes := request[COMMAND_LENGTH:]
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
	commandBytes := request[COMMAND_LENGTH:]
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

func handleTransactions(request []byte, bc *Blockchain) {
	commandBytes := request[COMMAND_LENGTH:]
	var txs []*Transaction
	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&txs)
	if err != nil {
		log.Panic(err)
	}
	sendTransactionToMiner(knowNodes[1], txs)
}

func handleRequireMine(request []byte, bc *Blockchain) {
	commandBytes := request[COMMAND_LENGTH:]
	var txs []*Transaction
	decoder := gob.NewDecoder(bytes.NewReader(commandBytes))

	err := decoder.Decode(&txs)
	if err != nil {
		log.Panic(err)
	}
	nodeID := os.Getenv("NODE_ID")
	txp := NewTXPool(nodeID)
	//将txs保存到交易池
	txp.Txs = append(txp.Txs, txs...)
	txp.saveFile(nodeID)

	const packageNum = 1

	//判断交易池是否有足够的交易
	if len(txp.Txs) > 0 {
		//开始挖矿
		fmt.Println("开始挖矿")
		blockchain := BlockchainObject(nodeID)
		//取出要打包的交易
		newBlock := blockchain.MineNewBlock(txs)
		txp.Txs = txp.Txs[packageNum:]
		txp.saveFile(nodeID)
		//发送newBlock 给主节点验证工作量证明
		sendNewBlockToMain(knowNodes[0], newBlock)
	}
}

func handleVerifyBlock(request []byte, blockchain *Blockchain) {
	commandBytes := request[COMMAND_LENGTH:]
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
		sendVersion(knowNodes[1], blockchain)
	}

}

func sendData(to string, data []byte) {
	conn, err := net.Dial("tcp", to)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
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


func sendGetData(to string, kind string, hash []byte) {
	getData := GetData{nodeAddress, kind, hash}
	sendCommandData(COMMAND_GETDATA, getData, to)
}

func sendBlock(to string, block *Block) {
	blockData := BlockData{nodeAddress, gobEncode(block)}
	sendCommandData(COMMAND_BLOCKDATA, blockData, to)
}

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
	payload := gobEncode(data)
	request := append(commandToBytes(command), payload...)
	sendData(to, request)
}
