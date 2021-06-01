package main

import (
	"./BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()
	//blc := BLC.BlockchainObject()
	//unUTXO := blc.UnUTXOs("liyuechun")
	//for i := 0; i < len(unUTXO); i++ {
	//	fmt.Println(unUTXO[i])
	//}
}
