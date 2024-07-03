package main

import (
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg_samples"
)

func main() {
	count := 0
	errs := 0

	fn := func() {
		serverPrivateKey, _ := utils.GeneratePrivateKey()
		server := xchg_samples.NewSimpleServer(serverPrivateKey)
		server.Start()

		serverAddress := utils.AddressForPublicKey(&serverPrivateKey.PublicKey)
		fmt.Println(serverAddress)
		client := xchg_samples.NewSimpleClient(serverAddress)

		for {
			time.Sleep(100 * time.Millisecond)
			var err error
			v, err := client.Version()
			fmt.Println("Result", v)
			if err != nil {
				errs++
			} else {
				count++
			}
		}
	}

	for i := 0; i < 1; i++ {
		go fn()
		time.Sleep(100 * time.Millisecond)
	}

	for {
		time.Sleep(1 * time.Second)
		//fmt.Println("res:", count, errs)
		count = 0
		errs = 0
	}
}
