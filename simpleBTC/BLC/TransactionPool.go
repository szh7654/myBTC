package BLC

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type TransactionPool struct {
	Txs []*Transaction
}

func NewTXPool(nodeID string) *TransactionPool {
	txPollFile := fmt.Sprintf(txPollFile, nodeID)
	if _, err := os.Stat(txPollFile); os.IsNotExist(err) {
		fmt.Println("交易池不存在。。。创建交易池")
		txp := &TransactionPool{[]*Transaction{}}
		return txp
	}

	txpBytes, err := ioutil.ReadFile(txPollFile)
	if err != nil {
		log.Panic(err)
	}

	var txp TransactionPool

	reader := bytes.NewReader(txpBytes)
	decoder := gob.NewDecoder(reader)
	err = decoder.Decode(&txp)
	if err != nil {

		log.Panic(err)
	}
	return &txp
}

func (txp *TransactionPool) saveFile(nodeID string) {
	txPollFile := fmt.Sprintf(txPollFile, nodeID)
	txpBytes := gobEncode(txp)
	err := ioutil.WriteFile(txPollFile, txpBytes, 0644)
	if err != nil {
		log.Panic(err)
	}
}
