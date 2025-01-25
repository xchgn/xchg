package main

import (
	"encoding/json"
	"fmt"

	"github.com/xchgn/xchg/blockchain"
)

func main() {
	bl := blockchain.NewBlockchain()
	for i := 0; i < 4; i++ {
		n, err := bl.GetNetwork(i)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		nBS, _ := json.MarshalIndent(n, "", "  ")
		fmt.Println(string(nBS))
	}
}
