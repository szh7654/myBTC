package main

import (
	"crypto/sha256"
)

type Transaction struct {
	TxHash [32]byte      //交易hash
	Vins   []*TXInput  //输入
	Vouts  []*TXOutput // 输出
}

func (tx *Transaction) NewCoinbaseTransaction(addr string) *Transaction {
	txInput := &TXInput{[32]byte{}, -1, "Genesis Data"}
	txOutput := &TXOutput{10, addr}
	txCoinbase := &Transaction{[32]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}

	tx.TxHash = sha256.Sum256(ToBytes(tx))

	return txCoinbase

}

