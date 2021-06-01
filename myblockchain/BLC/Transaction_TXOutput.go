package BLC

type TXOutput struct {
	Value int64
	ScriptPubKey string
}

// 解锁
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	return txOutput.ScriptPubKey == address
}