package samplesinglecall

import (
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

func Run() {
	serverPrivateKey, _ := utils.GeneratePrivateKey()
	s := xchg.StartServerPeer(serverPrivateKey, func(param *xchg.Param) (response []byte, err error) {
		response = []byte("DATA")
		return
	})

	c := xchg.StartClientPeer()
	resultBS, err := c.Call(s.Address(), "", "", nil, 2*time.Second)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Success:", string(resultBS))

	c.Stop()
	s.Stop()
}
