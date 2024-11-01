package xchg

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func (c *Peer) getFramesFromRouters() {
	c.mtx.Lock()
	network := c.network
	c.mtx.Unlock()

	if network == nil {
		return
	}

	addr := network.GetRouterAddr(hex.EncodeToString(c.Address()))
	c.getFramesFromRouter(addr)
}

func (c *Peer) getFramesFromRouter(router string) {
	c.mtx.Lock()
	processing := c.gettingFromInternet[router]
	c.mtx.Unlock()
	if processing {
		return
	}
	c.mtx.Lock()
	c.gettingFromInternet[router] = true
	c.mtx.Unlock()
	defer func() {
		c.mtx.Lock()
		c.gettingFromInternet[router] = false
		c.mtx.Unlock()
	}()

	c.mtx.Lock()
	c.routerStatRead[router]++
	c.mtx.Unlock()

	{
		c.mtx.Lock()
		fromMessageId := c.lastReceivedMessageId[router]
		c.mtx.Unlock()
		getMessageRequest := make([]byte, 8+8+32)
		binary.LittleEndian.PutUint64(getMessageRequest[0:], fromMessageId)
		binary.LittleEndian.PutUint64(getMessageRequest[8:], 10*1024*1024)
		copy(getMessageRequest[16:], c.localAddressBS)
		//logger.Println("GETTING .......................", hex.EncodeToString(c.Address())[:8])
		res, err := c.httpCall(c.httpClientLong, router, "r", getMessageRequest)
		//logger.Println("GETTING .......................OK", hex.EncodeToString(c.Address())[:8])
		if err != nil {
			fmt.Println("HTTP Error: ", err)
			return
		}

		if len(res) >= 8 {
			lastReceivedMessageId := binary.LittleEndian.Uint64(res[0:])
			c.mtx.Lock()
			c.lastReceivedMessageId[router] = lastReceivedMessageId
			c.mtx.Unlock()
			go c.processFramesFromInternet(res, router)
		}
	}
}

func (c *Peer) processFramesFromInternet(res []byte, router string) {
	offset := 8

	//fmt.Println("processFramesFromInternet", res)

	responses := make([]*Transaction, 0)
	responsesCount := 0

	framesCount := 0

	for offset < len(res) {
		if offset+69 <= len(res) {
			frameLen := int(binary.LittleEndian.Uint32(res[offset:]))
			if offset+frameLen <= len(res) {
				framesCount++
				//logger.Println("RCV:", utils.TransactionSummary(res[offset:offset+frameLen]))
				responseFrames := c.processFrame(router, res[offset:offset+frameLen])
				responses = append(responses, responseFrames...)
				responsesCount += len(responseFrames)
			} else {
				break
			}
			offset += frameLen
		} else {
			break
		}
	}

	if len(responses) > 0 {
		for _, f := range responses {
			addr := c.network.GetRouterAddr(f.DestAddressString())
			frame := f.Marshal()
			go c.httpCall(c.httpClient, addr, "w", frame)
		}
	}
}
