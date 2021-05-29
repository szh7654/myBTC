package main

type TXInput struct {
	TxHash      [32]byte // 交易的Hash
	Vout      int // 存储TXOutput在Vout里面的索引
	ScriptSig string // 输入的签名，即付款者
}