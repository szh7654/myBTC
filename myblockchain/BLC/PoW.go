package BLC

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// 前16位为0
const targetBit = 16

type PoW struct {
	Block  *Block
	target *big.Int //保证hash < target
}
// Return a PoW object
func NewPoW(block *Block) *PoW {
	target := big.NewInt(1)
	target = target.Lsh(target, 256-targetBit)
	return &PoW{block, target}
}

func (pow *PoW) Run() ([]byte, int64) {
	var hashInt big.Int
	var hash [32]byte
	for nonce := 0; ; nonce++ {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])
		// target > hashInt
		if pow.target.Cmp(&hashInt) == 1 {
			return hash[:], int64(nonce)
			break
		}
	}
	return nil, 0
}

// 准备要进行hash的data
func (pow *PoW) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			IntToHex(pow.Block.Timestamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
			IntToHex(int64(pow.Block.Height)),
		},
		[]byte{},
	)

	return data
}
