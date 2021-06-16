package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey 
	PublickKey []byte           


func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()

	return &Wallet{privateKey, publicKey}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

func (w *Wallet) GetAddress() []byte {
	pubKeyHash := PubKeyHash(w.PublickKey)
	address := PublicHashToAddress(pubKeyHash)
	return address
}

func PubKeyHash(publickKey []byte) []byte {
	hasher := sha256.New()
	hasher.Write(publickKey)
	hash1 := hasher.Sum(nil)
	hasher2 := ripemd160.New()
	hasher2.Write(hash1)
	hash2 := hasher.Sum(nil)
	return hash2
}


func CheckSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:addressCheckSumLen]
}


func IsValidAddress(address []byte) bool {
	full_payload := Base58Decode(address)
	checkSumBytes := full_payload[len(full_payload)-addressCheckSumLen:]
	version_payload := full_payload[:len(full_payload)-addressCheckSumLen]
	checkSumBytes2 := CheckSum(version_payload)
	return bytes.Compare(checkSumBytes, checkSumBytes2) == 0
}

func PublicHashToAddress(pubKeyHash []byte) []byte {
	versioned_payload := append([]byte{version}, pubKeyHash...)
	checkSumBytes := CheckSum(versioned_payload)
	full_payload := append(versioned_payload, checkSumBytes...)
	address := Base58Encode(full_payload)
	return address
}
