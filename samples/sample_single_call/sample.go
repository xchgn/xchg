package samplesinglecall

import (
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

func Run() {
	serverPrivateKey, _ := utils.GeneratePrivateKey()
	s := xchg.NewPeer(serverPrivateKey, xchg.NewDefaultLogger())
	s.ServerProcessorAuth = func(authData []byte) (err error) {
		return nil
	}
	counter := 0
	s.ServerProcessorCall = func(authData []byte, function string, parameter []byte) (response []byte, err error) {
		counter++
		return []byte("RESULT_" + fmt.Sprint(counter)), nil
	}
	s.Start(true)

	c := xchg.NewPeer(nil, xchg.NewDefaultLogger())
	c.Start(false)

	for i := 0; i < 10; i++ {

		resultBS, err := c.Call(s.Address(), "pass", "version", nil, 2*time.Second)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println("Success:", string(resultBS))
	}
	c.Stop()
	s.Stop()
}
