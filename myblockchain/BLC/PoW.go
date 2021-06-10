package BLC

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	Block  *Block   // The block which will be calculate
	target *big.Int // a block hash should satisfy hash < target
}

func PoWFactory(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target = target.Lsh(target, 256-targetBit)
	return &ProofOfWork{block, target}
}

func (pow *ProofOfWork) Run() ([]byte, int64) {
	nonce := 0
	var hashInt big.Int
	var hash [32]byte
	dataBytes := pow.prepareData()

	for {
		dataBytes := bytes.Join(
			[][]byte{ //[]byte的切片
				dataBytes,
				IntToHex(int64(nonce)),
			},
			[]byte{},
		)
		hash = sha256.Sum256(dataBytes)
		hashInt.SetBytes(hash[:])
		if pow.target.Cmp(&hashInt) == 1 {
			fmt.Printf("\nhash: %x\n", hash) //hash: 00ea9e3743900b6086acbb86390457f72fb3a4908609bd900536064f8e89448d
			break
		}
		nonce = nonce + 1
	}
	return hash[:], int64(nonce)
}

func (pow *ProofOfWork) prepareData() []byte {
	data := bytes.Join([][]byte{
		pow.Block.PrevBlockHash,
		pow.Block.HashTransactions(),
		IntToHex(pow.Block.Timestamp),
		IntToHex(int64(targetBit)),
		IntToHex(int64(pow.Block.Height)),
	}, []byte{},
	)

	return data
}

func (pow *ProofOfWork) IsValid() bool {
	var hashInt big.Int
	hashInt.SetBytes(pow.Block.BlockHash)
	if pow.target.Cmp(&hashInt) == 1 {
		return true
	}
	return false
}
