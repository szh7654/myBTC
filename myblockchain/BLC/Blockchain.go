package BLC

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"
)

// 数据库名字
const dbName = "blockchain.db"

// 表的名字
const blockTableName = "blocks"

type Blockchain struct {
	Tip []byte //最新的区块的Hash
	DB  *bolt.DB
}

// 迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

// 判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

// 遍历输出所有区块的信息
func (blc *Blockchain) Printchain() {

	fmt.Println("PrintchainPrintchainPrintchainPrintchain")
	blockchainIterator := blc.Iterator()

	for {
		block := blockchainIterator.Next()

		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Txs {

			fmt.Printf("tx.TxHash: %x\n", tx.TxHash)
			fmt.Println("Vins:")
			for _, in := range tx.Vins {
				fmt.Printf("in.TxHash: %x\n", in.TxHash)
				fmt.Printf("in.Vout: %d\n", in.Vout)
				fmt.Printf("in.ScriptSig: %s\n", in.ScriptSig)
			}

			fmt.Println("Vouts:")
			for _, out := range tx.Vouts {
				fmt.Println(out.Value)
				fmt.Println(out.ScriptPubKey)
			}
		}

		fmt.Println("------------------------------")

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

}

//// 增加区块到区块链里面
func (blc *Blockchain) AddBlockToBlockchain(txs []*Transaction) {

	err := blc.DB.Update(func(tx *bolt.Tx) error {

		//1. 获取表
		b := tx.Bucket([]byte(blockTableName))
		//2. 创建新区块
		if b != nil {

			// ⚠️，先获取最新区块
			blockBytes := b.Get(blc.Tip)
			// 反序列化
			block := DeserializeBlock(blockBytes)

			//3. 将区块序列化并且存储到数据库中
			newBlock := NewBlock(txs, block.Height+1, block.Hash)
			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//4. 更新数据库里面"l"对应的hash
			err = b.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			//5. 更新blockchain的Tip
			blc.Tip = newBlock.Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// 创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock(address string) *Blockchain {

	// 判断数据库是否存在
	if DBExists() {
		fmt.Println("创世区块已经存在.......")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块.......")

	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var genesisHash []byte

	// 关闭数据库
	err = db.Update(func(tx *bolt.Tx) error {

		// 创建数据库表
		b, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if b != nil {
			// 创建创世区块
			// 创建了一个coinbase Transaction
			txCoinbase := NewCoinbaseTransaction(address)

			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})
			// 将创世区块存储到表中
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			// 存储最新的区块的hash
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			genesisHash = genesisBlock.Hash
		}

		return nil
	})

	return &Blockchain{genesisHash, db}

}

// 返回Blockchain对象
func BlockchainObject() *Blockchain {

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			// 读取最新区块的Hash
			tip = b.Get([]byte("l"))

		}

		return nil
	})

	return &Blockchain{tip, db}
}


func (blockchain *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	fmt.Println(from)
	fmt.Println(to)
	fmt.Println(amount)
	var txs []*Transaction
	for index, address := range from {
		value, _ := strconv.Atoi(amount[0])
		tx := NewSimpleTransaction(address, to[index], value, blockchain, txs)
		txs = append(txs, tx)
		fmt.Println(tx)
	}
	var block *Block

	blockchain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte("l"))

			blockBytes := b.Get(hash)

			block = DeserializeBlock(blockBytes)

		}

		return nil
	})

	//2. 建立新的区块
	block = NewBlock(txs, block.Height+1, block.Hash)

	//将新区块存储到数据库
	blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			b.Put(block.Hash, block.Serialize())

			b.Put([]byte("l"), block.Hash)

			blockchain.Tip = block.Hash

		}
		return nil
	})

}

// 返回所有未花费的输出
func (blockchain *Blockchain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO
	// 遍历Vins，即记录该address所有的花销
	// key: Transaction hash
	// value: index if Vins 从而找到一个vin
	spentTXOutputs := make(map[string][]int)

	for _, tx := range txs{
		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {
				if in.UnLockWithAddress(address) {
					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _,tx := range txs {

	Work1:
		for index,out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				if len(spentTXOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash,indexArray := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if hash == txHashStr {
							var isUnSpentUTXO bool
							for _,outIndex := range indexArray {
								if index == outIndex {
									isUnSpentUTXO = true
									continue Work1
								}
								if isUnSpentUTXO == false {
									utxo := &UTXO{tx.TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}


	blockIterator := blockchain.Iterator()
	for {
		// 因为一个区块的Input所对应的Output总会在这个区块前面，
		// 所以先后遍历每个区块的Vins和Vouts, 通过vins过滤掉花费过的Outputs
		block := blockIterator.Next()
		//if block != nil {
		for i := len(block.Txs) - 1; i >= 0 ; i-- {
			tx := block.Txs[i]
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Vins {
					if in.UnLockWithAddress(address) {
						key := hex.EncodeToString(in.TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}
				}
			}
		work:
			for index, out := range tx.Vouts {
				if out.UnLockScriptPubKeyWithAddress(address) {
					if spentTXOutputs != nil {
						if len(spentTXOutputs) != 0 {
							var isSpentUTXO bool
							for txHash, indexArray := range spentTXOutputs {
								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}
							if isSpentUTXO == false {
								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
		//} else {
		//	break
		//}

	}
	return unUTXOs
}

// 查询余额
func (blockchain *Blockchain) GetBalance(address string) int64 {
	utxos := blockchain.UnUTXOs(address, []*Transaction{})
	fmt.Println("===================")
	for _,utxo := range utxos {
		fmt.Println(utxo)
	}
	var amount int64

	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}
	return amount
}

func (blockchain *Blockchain) FindSpendableUTXOS(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	utxos := blockchain.UnUTXOs(from, txs)
	spendableUTXO := make(map[string][]int)
	var value int64
	for _, utxo := range utxos {
		value = value + utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Printf("%s's fund is 不足\n", from)
		os.Exit(1)
	}
	return value, spendableUTXO
}
