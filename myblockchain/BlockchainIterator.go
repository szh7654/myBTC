package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	CurHash []byte
	DB      *bolt.DB
}

// 返回上一个区块
func (iter *BlockchainIterator) Prev() *Block {
	var block *Block
	err := iter.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			block = DeSerialize(b.Get(iter.CurHash))
			iter.CurHash = block.PrevBlockHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}
