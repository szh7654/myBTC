package BLC

import (
	"fmt"
	"time"
)

type Block struct {
	Height        int64          // Height of current Block
	PrevBlockHash []byte         // BlockHash of previous Block
	Transactions  []*Transaction // Transactions in this Block
	Timestamp     int64          // Timestamp when the block is packaged
	BlockHash     []byte         // BlockHash of current Block
	Nonce         int64          // Nonce of current Block
}

// Mine a new Block
func NewBlock(txs []*Transaction, height int64, preBlockHash []byte) *Block {
	block := &Block{height, preBlockHash, txs, time.Now().Unix(), nil, 0}

	// Proof of Work
	pow := PoWFactory(block)
	hash, nonce := pow.Run()
	block.BlockHash = hash[:]
	block.Nonce = nonce

	return block
}

func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, 1, make([]byte, 32, 32))
}

// Return the hash of Transactions in current Block
func (block *Block) HashTransactions() []byte {
	//将txs的hash序列化为[]byte,并放进一个数组里面
	var txs [][]byte
	for _, tx := range block.Transactions {
		txBytes := gobEncode(tx)
		txs = append(txs, txBytes)
	}
	// Calculate the root hash of merkle tree in current block
	merkleTree := NewMerkleTree(txs)
	return merkleTree.root.data
}

func (block *Block) String() string {
	return fmt.Sprintf(
		"\n------------------------------"+
			"\nBlock's Info:\n\t"+
			"height:%d,\n\t"+
			"PreHash:%x,\n\t"+
			"Transactions: %v,\n\t"+
			"Timestamp: %s,\n\t"+
			"BlockHash: %x,\n\t"+
			"Nonce: %v\n\t",
		block.Height,
		block.PrevBlockHash,
		block.Transactions,
		time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"),
		block.BlockHash, block.Nonce)
}
