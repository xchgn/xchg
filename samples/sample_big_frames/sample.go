package samplebigframes

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

var data []byte

func Run() {
	data = make([]byte, xchg.XchgMaxTransactionSize)
	rand.Read(data[:])

	hash := sha256.New()
	hash.Write(data)
	fmt.Println("Data Hash:", hex.EncodeToString(hash.Sum(nil)))

	serverPrivateKey, _ := utils.GeneratePrivateKey()
	s := xchg.StartServerPeer(serverPrivateKey, func(param *xchg.Param) (response []byte, err error) {
		if param.Function == "" {
			return nil, nil
		}
		response = data
		return
	})

	///////////////////////////////////////////////
	// Make client
	c := xchg.StartClientPeer()
	for i := 0; i < 10; i++ {
		resultBS, err := c.Call(s.Address(), "", "data", nil, 2*time.Second)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		h := sha256.New()
		h.Write(resultBS)
		receivedHash := h.Sum(nil)
		fmt.Println(i, "Received Data Hash:", hex.EncodeToString(receivedHash))

		//fmt.Println("Success:", resultBS[:4], len(resultBS))
	}

	c.Stop()
	s.Stop()
}
