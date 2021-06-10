package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

type Transaction struct {
	TransactionHash []byte    // 交易hash
	Inputs          []*Input  // 输入
	Outputs         []*Output // 输出
}

func NewCoinbaseTransacion(address string) *Transaction {
	input := &Input{[]byte{}, -1, nil, nil}
	output := NewOutput(10, address)
	txCoinBaseTransaction := &Transaction{[]byte{}, []*Input{input}, []*Output{output}}
	txCoinBaseTransaction.SetTransaction()
	return txCoinBaseTransaction
}

func NewRewardTransacion() *Transaction {
	input := &Input{[]byte{}, -1, nil, nil}
	address := CoinbaseAddress(os.Getenv("NODE_ID"))
	if address == "" {
		wallets := NewWallets(os.Getenv("NODE_ID"))
		for walletAddress := range wallets.WalletMap {
			address = walletAddress
		}
	}
	if address == "" {
		log.Panic("未定义地址")
	}
	// give 1 BTC as a reward
	output := NewOutput(1, address)
	txCoinBaseTransaction := &Transaction{[]byte{}, []*Input{input}, []*Output{output}}
	txCoinBaseTransaction.SetTransaction()
	return txCoinBaseTransaction
}

// 普通交易产生的Transaction
func NewSimpleTransation(from string, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction, nodeID string) *Transaction {
	var inputs []*Input
	var outputs []*Output

	total, spendableUTXO := utxoSet.FindSpendableUTXOs(from, amount, txs)

	wallets := NewWallets(nodeID)
	wallet := wallets.WalletMap[from]
	if wallet == nil {
		log.Panic("钱包错误")
	}

	for transactionHash, indexArray := range spendableUTXO {
		transactionHashBytes, _ := hex.DecodeString(transactionHash)
		for _, index := range indexArray {
			input := &Input{transactionHashBytes, index, nil, wallet.PublickKey}
			inputs = append(inputs, input)
		}
	}

	output := NewOutput(amount, to)
	outputs = append(outputs, output)

	// 找零
	output2 := NewOutput(total-amount, from)
	outputs = append(outputs, output2)

	tx := &Transaction{[]byte{}, inputs, outputs}
	tx.SetTransaction()
	// 签名
	utxoSet.blockChain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}

func (tx *Transaction) IsCoinBaseTransaction() bool {
	return len(tx.Inputs[0].TransactionHash) == 0 && tx.Inputs[0].IndexOfOutputs == -1
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTransactionMap map[string]*Transaction) {
	if tx.IsCoinBaseTransaction() {
		return
	}

	for _, input := range tx.Inputs {
		if prevTransactionMap[hex.EncodeToString(input.TransactionHash)] == nil {
			log.Panic("无法签名。")
		}
	}

	txCopy := tx.TrimmedCopy()

	for index, input := range txCopy.Inputs {
		prevTx := prevTransactionMap[hex.EncodeToString(input.TransactionHash)]
		txCopy.Inputs[index].Signature = nil
		txCopy.Inputs[index].PublicKey = prevTx.Outputs[input.IndexOfOutputs].PubKeyHash
		txCopy.TransactionHash = txCopy.NewTransactionHash()
		txCopy.Inputs[index].PublicKey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.TransactionHash)
		if err != nil {
			log.Panic(err)
		}

		sign := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[index].Signature = sign
	}

}

func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []*Input
	var outputs []*Output
	for _, in := range tx.Inputs {
		inputs = append(inputs, &Input{in.TransactionHash, in.IndexOfOutputs, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, &Output{out.Value, out.PubKeyHash})
	}

	txCopy := &Transaction{[]byte{}, inputs, outputs}
	return txCopy

}

func (tx *Transaction) SetTransaction() {
	txBytes := gobEncode(tx)
	allBytes := bytes.Join([][]byte{txBytes, IntToHex(time.Now().Unix())}, []byte{})
	hash := sha256.Sum256(allBytes)
	tx.TransactionHash = hash[:]
}

func (tx *Transaction) NewTransactionHash() []byte {
	txCopy := tx
	txCopy.TransactionHash = []byte{}
	txBytes := gobEncode(txCopy)
	hash := sha256.Sum256(txBytes)
	return hash[:]
}

func (tx *Transaction) VerifyTransaction(prevTransactionMap map[string]*Transaction) bool {
	if tx.IsCoinBaseTransaction() {
		return true
	}

	for _, input := range tx.Inputs { //
		if prevTransactionMap[hex.EncodeToString(input.TransactionHash)] == nil {
			log.Panic("无法验证")
		}
	}

	txCopy := tx.TrimmedCopy()

	curev := elliptic.P256()
	for index, input := range tx.Inputs {
		// calculate hash value
		prevTx := prevTransactionMap[hex.EncodeToString(input.TransactionHash)]
		txCopy.Inputs[index].Signature = nil
		txCopy.Inputs[index].PublicKey = prevTx.Outputs[input.IndexOfOutputs].PubKeyHash
		txCopy.TransactionHash = txCopy.NewTransactionHash()
		txCopy.Inputs[index].PublicKey = nil
		// get public key
		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PublicKey)
		x.SetBytes(input.PublicKey[:keyLen/2])
		y.SetBytes(input.PublicKey[keyLen/2:])
		rawPublicKey := ecdsa.PublicKey{curev, &x, &y}
		// get signature
		r := big.Int{}
		s := big.Int{}
		signLen := len(input.Signature)
		r.SetBytes(input.Signature[:signLen/2])
		s.SetBytes(input.Signature[signLen/2:])
		// verify
		if ecdsa.Verify(&rawPublicKey, txCopy.TransactionHash, &r, &s) == false {
			fmt.Println("验证失败")
			return false
		}
	}
	return true
}

func (tx *Transaction) String() string {
	var vinStrings [][]byte
	for _, vin := range tx.Inputs {
		vinString := fmt.Sprint(vin)
		vinStrings = append(vinStrings, []byte(vinString))
	}
	vinString := bytes.Join(vinStrings, []byte{})

	var outStrings [][]byte
	for _, out := range tx.Outputs {
		outString := fmt.Sprint(out)
		outStrings = append(outStrings, []byte(outString))
	}

	outString := bytes.Join(outStrings, []byte{})

	return fmt.Sprintf("\n\r\t\t===============================\n\r\t\tTxID: %x, \n\t\tVins: %v, \n\t\tVout: %v\n\t\t", tx.TransactionHash, string(vinString), string(outString))
}

type Input struct {
	TransactionHash []byte //  交易的Hash
	IndexOfOutputs  int
	Signature       []byte //数字签名
	PublicKey       []byte //钱包里的公钥
}

func (input *Input) UnlockWithAddress(pubKeyHash []byte) bool {
	pubKeyHash2 := PubKeyHash(input.PublicKey)
	return bytes.Compare(pubKeyHash, pubKeyHash2) == 0
}

func (input *Input) String() string {
	return fmt.Sprintf("\n\t\t\tTxInput_TXID: %x, IndexOfOutputs: %v, Signature: %x, PublicKey:%x", input.TransactionHash, input.IndexOfOutputs, input.Signature, input.PublicKey)
}

type Output struct {
	Value      int64  //金额
	PubKeyHash []byte //公钥哈希
}

func (output *Output) UnlockWithAddress(address string) bool {
	full_payload := Base58Decode([]byte(address))
	pubKeyHash := full_payload[1 : len(full_payload)-addressCheckSumLen]
	return bytes.Compare(pubKeyHash, output.PubKeyHash) == 0
}

func NewOutput(value int64, address string) *Output {
	output := &Output{value, nil}
	output.Lock(address)
	return output
}

func (output *Output) Lock(address string) {
	full_payload := Base58Decode([]byte(address))
	//获取公钥hash
	output.PubKeyHash = full_payload[1 : len(full_payload)-addressCheckSumLen]
}

func (output *Output) String() string {
	return fmt.Sprintf("\n\t\t\tValue: %d, PubKeyHash(转成地址显示): %s", output.Value, PublicHashToAddress(output.PubKeyHash))
}

type UTXO struct {
	TransactionHash []byte  // 该Output所在的hash
	Index           int     // 该Output 的下标
	Output          *Output // Output
}

func (utxo *UTXO) String() string {
	return fmt.Sprintf(
		"\n------------------------------"+
			"\nA UTXO's Info:\n\t"+
			"TransactionHash:%s,\n\t"+
			"Index:%d,\n\t"+
			"Output: %v,\n\t",
		hex.EncodeToString(utxo.TransactionHash),
		utxo.Index,
		utxo.Output,
	)
}