package BLC

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type BlockchainIterator struct {
	currentHash []byte   //当前hash
	DB          *bolt.DB //数据库
}

func (blockchainIterator *BlockchainIterator) Next() *Block {
	DBName := fmt.Sprintf(DBName, os.Getenv("NODE_ID"))
	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var block Block
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucketName))
		if b != nil {
			currentBlockBytes := b.Get(blockchainIterator.currentHash)
			gobDecode(currentBlockBytes, &block)
			blockchainIterator.currentHash = block.PrevBlockHash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &block
}
