package BLC

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type Block struct {
	Height        int64          // Height of the block
	PrevBlockHash []byte         // Block Hash of Previous Block
	Txs           []*Transaction // Transaction Array of the block
	Timestamp     int64          // Created time
	Hash          []byte         // Hash of the Block
	Nonce         int64
}

// Convert Transactions to byte
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	// Combine hashes of Transactions
	for _, tx := range block.Txs {
		txHashes = append(txHashes, tx.TxHash)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// Convert Block object to byte array
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// Convert byte array to Block object
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

// Create a new Block
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {
	block := &Block{height, prevBlockHash, txs, time.Now().Unix(), nil, 0}
	// Get a pow object
	pow := NewPoW(block)
	// Compute the nonce and corresponding hash
	hash, nonce := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Printf("A new Block Hash: %x\n", &hash)
	return block
}

//2. 单独写一个方法，生成创世区块

func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
