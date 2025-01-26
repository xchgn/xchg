package main

import (
	"encoding/hex"
	"fmt"

	"github.com/xchgn/xchg/blockchain"
)

func main() {
	ra := blockchain.NewRouterAccount()
	ra.GenerateCheques()
	routerObject, err := ra.FetchRouterObject()
	fmt.Println("Router object:", routerObject.IpAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	items := make([]blockchain.Item, 0)

	for i, ch := range routerObject.ChequeIds {
		if i > 30 {
			break
		}
		fmt.Println("ChequeId:", ch)
		cheque, _ := hex.DecodeString(ch[2:])
		items = append(items, blockchain.Item{ChequeId: cheque, AppId: []byte("app_id")})
	}

	ra.ApplyChecques(items)
}
