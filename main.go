package main

import (
	"fmt"

	"github.com/xchgn/xchg/blockchain"
)

func main() {
	ra := blockchain.NewRouterAccount()
	addr := ra.SuiClient().Account().Address
	fmt.Println("Xchg address:", addr)
}
