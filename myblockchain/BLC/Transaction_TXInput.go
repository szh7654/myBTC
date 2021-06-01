package BLC

// Corresponding TxOutput can be found through TxHash and Vout
type TXInput struct {
	TxHash    []byte // Hashcode of Transaction
	Vout      int    // Index of TXOutput in a transaction
	ScriptSig string // Payer address
}
// 判断当前的消费是谁的钱
func (txInput *TXInput) UnLockWithAddress(address string) bool {
	return txInput.ScriptSig == address
}