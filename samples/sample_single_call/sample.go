package samplesinglecall

import (
	"encoding/hex"
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

	fmt.Println("Server Address:", hex.EncodeToString(s.Address()))

	c := xchg.StartClientPeer()

	fmt.Println("Client Address:", hex.EncodeToString(c.Address()))

	resultBS, err := c.Call(s.Address(), "", "", nil, 2*time.Second)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Success:", string(resultBS))

	c.Stop()
	s.Stop()
}
