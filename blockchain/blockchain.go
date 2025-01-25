package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/xchgn/suigo/client"
)

func GetRouterForAddress(address string) string {
	return "localhost:8084"
}

type Blockchain struct {
	cl       *client.Client
	fundObj  *FundSuiObj
	networks map[int]*Network
}

func NewBlockchain() *Blockchain {
	var c Blockchain
	c.cl = client.NewClient(client.TESTNET_URL)
	c.networks = make(map[int]*Network)
	return &c
}

func (c *Blockchain) GetFundObject() (*FundSuiObj, error) {
	if c.fundObj != nil {
		return c.fundObj, nil
	}

	fmt.Println("Getting fund object")

	var showOptions client.GetObjectShowOptions
	showOptions.ShowContent = true
	objData, err := c.cl.GetObject(FUND_OBJECT_ID, showOptions)
	if err != nil {
		return nil, err
	}
	objDataBS, err := json.MarshalIndent(objData.Data.Content, "", "  ")
	if err != nil {
		return nil, err
	}

	var fundSuiObj FundSuiObj
	err = json.Unmarshal(objDataBS, &fundSuiObj)

	c.fundObj = &fundSuiObj

	return &fundSuiObj, nil
}

func (c *Blockchain) GetNetwork(segment int) (*Network, error) {
	if network, ok := c.networks[segment]; ok {
		return network, nil
	}

	fmt.Println("Getting network", segment)

	f, err := c.GetFundObject()
	if err != nil {
		return nil, err
	}

	net0, err := c.cl.GetDynamicFieldObject(f.Fields.Network.Fields.ID.ID, "u32", segment)
	if err != nil {
		return nil, err
	}
	net0BS, err := json.MarshalIndent(net0.Data.Content, "", "  ")
	if err != nil {
		return nil, err
	}

	var net0Obj NetworkSegment
	err = json.Unmarshal(net0BS, &net0Obj)
	if err != nil {
		return nil, err
	}

	var network Network
	network.Segment = net0Obj.Fields.Value.Fields.Index
	for _, router := range net0Obj.Fields.Value.Fields.Routers {
		stake, err := strconv.ParseUint(router.Fields.CurrentStake, 10, 64)
		if err != nil {
			continue
		}
		network.Routers = append(network.Routers, &Router{
			XchgAddr:     router.Fields.XchgAddress,
			CurrentStake: stake,
			IpAddr:       router.Fields.IpAddr,
		})
	}
	return &network, nil
}

func (c *Blockchain) Update() {

}
