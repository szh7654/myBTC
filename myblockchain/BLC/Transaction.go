package BLC

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
)

type Transaction struct {
	TxHash []byte      // Hashcode of a Transaction
	Vins   []*TXInput  // Inputs of a Transaction
	Vouts  []*TXOutput // Outputs of a Transaction
}

//1. 创世区块创建时的Transaction
func NewCoinbaseTransaction(address string) *Transaction {

	//代表消费
	txInput := &TXInput{[]byte{}, -1, "Genesis Data"}

	txOutput := &TXOutput{10, address}

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}

	//设置hash值
	txCoinbase.HashTransaction()

	return txCoinbase
}

func (tx *Transaction) HashTransaction() {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(result.Bytes())

	tx.TxHash = hash[:]
}

func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

// 转账时产生的Transaction
func NewSimpleTransaction(from string, to string, amount int, blockchain *Blockchain, txs []*Transaction) *Transaction {
	money, spendableUTXOdic := blockchain.FindSpendableUTXOS(from, amount, txs)
	var txInputs []*TXInput
	var txOutputs []*TXOutput
	for txHash, indexArray := range spendableUTXOdic {
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArray {
			txInput := &TXInput{txHashBytes, index, from}
			txInputs = append(txInputs, txInput)
		}
	}
	// 转账
	txOutput := &TXOutput{int64(amount),to}
	txOutputs = append(txOutputs,txOutput)

	// 找零
	txOutput = &TXOutput{int64(money) - int64(amount),from}
	txOutputs = append(txOutputs,txOutput)

	tx := &Transaction{[]byte{},txInputs,txOutputs}
	tx.HashTransaction()
	return tx
}
