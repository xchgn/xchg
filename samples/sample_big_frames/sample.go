package samplebigframes

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

func Run() {
	serverPrivateKey, _ := utils.GeneratePrivateKey()
	s := xchg.StartServerPeer(serverPrivateKey, func(param *xchg.Param) (response []byte, err error) {
		if param.Function == "" {
			return nil, nil
		}
		data := make([]byte, 10*1000000)
		rand.Read(data[:])
		response = data
		return
	})

	///////////////////////////////////////////////
	// Make client
	c := xchg.StartClientPeer()
	for i := 0; i < 1; i++ {
		resultBS, err := c.Call(s.Address(), "", "data", nil, 10*time.Second)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Success:", resultBS[:4], len(resultBS))
	}

	c.Stop()
	s.Stop()
}
