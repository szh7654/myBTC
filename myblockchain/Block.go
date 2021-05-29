package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Height        int64
	PrevBlockHash []byte
	Data          []byte
	Timestamp     int64
	Hash          []byte
	Nonce         int64
}

func (block *Block) Serialize() []byte {
	return ToBytes(block)
}

func DeSerialize(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	if err := decoder.Decode(&block); err != nil {
		log.Panic(err)
	}
	return &block
}

// new Block
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {
	block := &Block{height, prevBlockHash, []byte(data), time.Now().Unix(), nil, 0}
	pow := NewPoW(block)
	block.Hash, block.Nonce = pow.Run()
	return block
}

// create a genesis block
func CreateGenesisBlock(data string) *Block {
	return NewBlock(data, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
