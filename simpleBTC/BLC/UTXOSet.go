package BLC

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type UTXOSet struct {
	blockChain *Blockchain
}

/*
	查询block块中所有的未花费utxo：执行FindUnspentUTXOMap--->map
*/
func (utxoset *UTXOSet) ResetUTXOSet() {
	//blockchain.GetAllUTXOs 涉及到数据库, 现在要放在下面打开数据库的命令之前.不然卡死
	utxoMap := utxoset.blockChain.GetAllUTXOs()

	DBName := fmt.Sprintf(DBName, os.Getenv("NODE_ID"))
	fmt.Println(DBName)
	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		//1.utxoset表存在，删除
		b := tx.Bucket([]byte(UTXOSetBucketName))
		if b != nil {
			err := tx.DeleteBucket([]byte(UTXOSetBucketName))
			if err != nil {
				log.Panic(err)
			}
		}

		b, err := tx.CreateBucket([]byte(UTXOSetBucketName))
		if err != nil {
			log.Panic(err)
		}

		if b != nil {

			for transactionHash, outs := range utxoMap {
				transactionHash, _ := hex.DecodeString(transactionHash)
				b.Put(transactionHash, gobEncode(outs))
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}

/*
	查询对应地址的余额
*/
func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	var total int64

	utxos := utxoSet.FindPackedUTXO(address)

	for _, utxo := range utxos {
		total += utxo.Output.Value
	}

	return total
}

func (utxoSet *UTXOSet) FindPackedUTXO(address string) []*UTXO {
	var utxos []*UTXO
	DBName := fmt.Sprintf(DBName, os.Getenv("NODE_ID"))
	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOSetBucketName))
		if b != nil {
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				outputs := DeserializeTxOutputs(v)
				for _, utxo := range outputs.UTXOs {
					if utxo.Output.UnlockWithAddress(address) {
						utxos = append(utxos, utxo)
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return utxos
}

func (utxoSet *UTXOSet) FindSpendableUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	var total int64
	spendableUTXOMap := make(map[string][]int)

	unpackedUTXO := utxoSet.FindUnpackedUTXO(from, txs)
	for _, utxo := range unpackedUTXO {
		total += utxo.Output.Value
		transactionHash := hex.EncodeToString(utxo.TransactionHash)
		spendableUTXOMap[transactionHash] = append(spendableUTXOMap[transactionHash], utxo.Index)
		if total >= amount {
			return total, spendableUTXOMap
		}
	}

	packedUTXO := utxoSet.FindPackedUTXO(from)
	for _, utxo := range packedUTXO {
		total += utxo.Output.Value
		transactionHash := hex.EncodeToString(utxo.TransactionHash)
		spendableUTXOMap[transactionHash] = append(spendableUTXOMap[transactionHash], utxo.Index)
		if total >= amount {
			return total, spendableUTXOMap
		}
	}

	if total < amount {
		fmt.Printf("%s 的余额不足, 无法转账. 余额为: %d", from, total)
		os.Exit(1)
	}

	return total, spendableUTXOMap
}

func (utxoSet *UTXOSet) FindUnpackedUTXO(from string, txs []*Transaction) []*UTXO {
	var utxos []*UTXO
	spentTxOutputMap := make(map[string][]int)
	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]
		utxos = caculate(tx, from, spentTxOutputMap, utxos)
	}
	return utxos
}

/*
	更新数据库UTXO
*/
func (utxoSet *UTXOSet) Update() {
	//对最后一个区块进行处理
	lastBlock := utxoSet.blockChain.Iterator().Next()

	//遍历TXs 获取所有input
	inputs := []*Input{}
	for _, tx := range lastBlock.Transactions {
		if !tx.IsCoinBaseTransaction() {
			for _, input := range tx.Inputs {
				inputs = append(inputs, input)
			}
		}
	}

	//遍历TXs 获取UTXO
	outsMap := make(map[string]*UTXOArray)
	for _, tx := range lastBlock.Transactions {
		//每个交易中的utxo数组
		utxos := []*UTXO{}
		for outIndex, txOut := range tx.Outputs {
			isSpent := false
			for _, input := range inputs {
				if input.IndexOfOutputs == outIndex &&
					bytes.Compare(input.TransactionHash, tx.TransactionHash) == 0 {
					//已花费
					isSpent = true
					break
				}
			}
			if isSpent == false {
				utxo := &UTXO{tx.TransactionHash, outIndex, txOut}
				utxos = append(utxos, utxo)
			}
		}

		if len(utxos) > 0 {
			transactionHash := hex.EncodeToString(tx.TransactionHash)
			outputs := &UTXOArray{utxos}
			outsMap[transactionHash] = outputs
		}
	}

	//for txid, outputs := range outsMap {
	//	fmt.Printf("---------transactionHash :%s", txid)
	//	for _, utxo := range outputs.UTXOs {
	//		fmt.Println(utxo)
	//	}
	//}

	//获取utxo表,将input对应的utxo删除, 添加outsMap中的utxo
	DBName := fmt.Sprintf(DBName, os.Getenv("NODE_ID"))
	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOSetBucketName))
		if b != nil {
			//1.删除inputs对应的utxo
			for _, input := range inputs {
				outputsBytes := b.Get(input.TransactionHash)
				if len(outputsBytes) == 0 {
					continue
				}

				outputs := DeserializeTxOutputs(outputsBytes)

				//是否需要被删除
				isNeedDelete := false

				//将当前outputs里面未被消费的utxo 保存起来
				utxos := []*UTXO{}

				for _, utxo := range outputs.UTXOs {
					if bytes.Compare(utxo.TransactionHash, input.TransactionHash) == 0 &&
						input.IndexOfOutputs == utxo.Index &&
						input.UnlockWithAddress(utxo.Output.PubKeyHash) {
						//已花费
						isNeedDelete = true
						continue
					}

					utxos = append(utxos, utxo)
				}

				if isNeedDelete {
					err := b.Delete(input.TransactionHash)
					if err != nil {
						log.Panic(err)
					}

					if len(utxos) > 0 {
						outputs := &UTXOArray{utxos}
						b.Put(input.TransactionHash, gobEncode(outputs))
					}
				}

			}

			//2.添加outsMap 到数据库中
			for transactionHash, outputs := range outsMap {
				transactionHashBytes, _ := hex.DecodeString(transactionHash)
				b.Put(transactionHashBytes, gobEncode(outputs))
			}

		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}
