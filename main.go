package main

import (
	"fmt"

	"github.com/xchgn/xchg/blockchain"
)

func main() {
	ra := blockchain.NewRouterAccount()
	_, err := ra.FetchRouterObject()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//ra.GenerateCheques()
	//fmt.Println("Xchg address:", addr)
}
