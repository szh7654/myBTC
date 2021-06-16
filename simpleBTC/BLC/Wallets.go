package BLC

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Wallets struct {
	WalletMap map[string]*Wallet
}


func NewWallets(nodeID string) *Wallets {
	walletsFile := fmt.Sprintf(walletsFile, nodeID)
	if _, err := os.Stat(walletsFile); os.IsNotExist(err) {
		fmt.Println("钱包文件不存在。。。")
		wallets := &Wallets{}
		wallets.WalletMap = make(map[string]*Wallet)
		return wallets
	}

	wsBytes, err := ioutil.ReadFile(walletsFile)
	if err != nil {
		log.Panic(err)
	}

	gob.Register(elliptic.P256())
	var wallets Wallets

	reader := bytes.NewReader(wsBytes)
	decoder := gob.NewDecoder(reader)
	err = decoder.Decode(&wallets)
	if err != nil {

		log.Panic(err)
	}

	return &wallets
}

func (ws *Wallets) CreateWallet(nodeID string) {
	wallet := NewWallet()
	address := wallet.GetAddress()

	fmt.Printf("创建的钱包地址：%s\n", address)

	ws.WalletMap[string(address)] = wallet

	ws.saveFile(nodeID)
}

func (ws *Wallets) saveFile(nodeID string) {
	walletsFile := fmt.Sprintf(walletsFile, nodeID)
	wsBytes := gobEncodeWithRegister(ws, elliptic.P256())
	err := ioutil.WriteFile(walletsFile, wsBytes, 0644)
	if err != nil {
		log.Panic(err)
	}
}
