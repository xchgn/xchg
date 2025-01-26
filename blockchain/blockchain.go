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
	fundObj  *MoveFundSuiObj
	networks map[int]*Network
}

func NewBlockchain() *Blockchain {
	var c Blockchain
	c.cl = client.NewClient(client.TESTNET_URL)
	c.networks = make(map[int]*Network)
	return &c
}

func (c *Blockchain) GetFundObject() (*MoveFundSuiObj, error) {
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

	var fundSuiObj MoveFundSuiObj
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

	var net0Obj MoveNetworkSegment
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

type RouterObject struct {
	Name             string
	Segment          int
	IpAddr           string
	XchgAddr         string
	Owner            string
	TotalStakeAmount uint64
}

func (c *Blockchain) GetRouterObject(routerXchgAddress string) (*RouterObject, error) {
	fundObj, err := c.GetFundObject()
	if err != nil {
		return nil, err
	}

	fmt.Println("---------------------------------")
	routerTableId := fundObj.Fields.Routers.Fields.ID.ID
	routerObj, err := c.cl.GetDynamicFieldObject(routerTableId, "address", routerXchgAddress)
	if err != nil {
		return nil, err
	}

	routerObjBS, err := json.MarshalIndent(routerObj.Data.Content, "", "  ")
	if err != nil {
		return nil, err
	}

	fmt.Println(string(routerObjBS))

	var routerObjStr MoveRouterObject

	err = json.Unmarshal(routerObjBS, &routerObjStr)
	if err != nil {
		return nil, err
	}

	stake, err := strconv.ParseUint(routerObjStr.Fields.Value.Fields.TotalStakeAmnt, 10, 64)
	if err != nil {
		return nil, err
	}

	var routerObject RouterObject
	routerObject.Name = routerObjStr.Fields.Value.Fields.Name
	routerObject.Segment = routerObjStr.Fields.Value.Fields.Segment
	routerObject.IpAddr = routerObjStr.Fields.Value.Fields.IpAddr
	routerObject.XchgAddr = "---"
	routerObject.Owner = routerObjStr.Fields.Value.Fields.Owner
	routerObject.TotalStakeAmount = stake

	fmt.Println("Router object:")
	fmt.Println("Name:", routerObject.Name)
	fmt.Println("Segment:", routerObject.Segment)
	fmt.Println("IpAddr:", routerObject.IpAddr)
	fmt.Println("XchgAddr:", routerObject.XchgAddr)
	fmt.Println("Owner:", routerObject.Owner)
	fmt.Println("TotalStakeAmount:", routerObject.TotalStakeAmount)

	// cheques
	fmt.Println("Cheque:")
	for i, cheque := range routerObjStr.Fields.Value.Fields.ChequeIds.Fields.Contents {
		fmt.Println("CH#", i, cheque)
	}

	fmt.Println("---------------------------------")

	return &routerObject, nil
}

func (c *Blockchain) Update() {

}
