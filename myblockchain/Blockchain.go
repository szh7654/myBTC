package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"time"
)

const db = "blockchain.db"

const bucketName = "blocks"

type Blockchain struct {
	Tip []byte //最新的区块的Hash
	DB  *bolt.DB
}

// 判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(db); os.IsNotExist(err) {
		return false
	}
	return true
}

// 从db中返回Blockchain对象
func BlockchainObject() *Blockchain {
	db, err := bolt.Open(db, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	return &Blockchain{tip, db}
}

// Create a blockchain with genesis block
func CreateBlockchainWithGenesisBlock(data string) {
	if DBExists() {
		fmt.Println("创世区块已经存在.......")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块.......")
	db, err := bolt.Open(db, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var blockHash []byte

	err = db.Update(func(tx *bolt.Tx) error {
		// 获取桶， 如果没有则创建
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			b, err = tx.CreateBucket([]byte(bucketName))
			if err != nil {
				log.Panic(err)
			}
		}

		if b != nil {
			genesisBlock := CreateGenesisBlock(data)

			// 存储创世区块
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 存储最新的区块的hash
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			blockHash = genesisBlock.Hash
		}
		return nil
	})
}

// Add a block to the blockchain
func (blockchain *Blockchain) AddBlockToBlockchain(data string) {
	// 更新db
	err := blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {

			// 获取最新block
			blockBytes := b.Get(blockchain.Tip)
			block := DeSerialize(blockBytes)

			// 创建block并存储
			newBlock := NewBlock(data, block.Height+1, block.Hash)
			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 更新db中的最新区块的hash
			err = b.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			// 更新链中的最新区块的hash
			blockchain.Tip = newBlock.Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

func (blockchain *Blockchain) Print() {
	iter := blockchain.Iterator()
	for {
		block := iter.Prev()

		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		fmt.Printf("Data：%s\n", block.Data)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nonce)
		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}
