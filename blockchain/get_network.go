package blockchain

// func GetFundObject(cl *client.Client) (*FundSuiObj, error) {
// 	var showOptions client.GetObjectShowOptions
// 	showOptions.ShowContent = true
// 	objData, err := cl.GetObject("0xfe53b68fbce0daa159c7abe893633584835bd398f9b6b8612a6d72fd72e9f1ff", showOptions)
// 	if err != nil {
// 		return nil, err
// 	}
// 	objDataBS, err := json.MarshalIndent(objData.Data.Content, "", "  ")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var fundSuiObj FundSuiObj
// 	err = json.Unmarshal(objDataBS, &fundSuiObj)

// 	return &fundSuiObj, nil
// }

// func GetNetwork(segment int) (*Network, error) {
// 	cl := client.NewClient(client.TESTNET_URL)
// 	f, _ := GetFundObject(cl)

// 	net0, err := cl.GetDynamicFieldObject(f.Fields.Network.Fields.ID.ID, "u32", segment)
// 	if err != nil {
// 		return nil, err
// 	}
// 	net0BS, err := json.MarshalIndent(net0.Data.Content, "", "  ")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var net0Obj NetworkSegment
// 	err = json.Unmarshal(net0BS, &net0Obj)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var network Network
// 	network.Segment = net0Obj.Fields.Value.Fields.Index
// 	for _, router := range net0Obj.Fields.Value.Fields.Routers {
// 		stake, err := strconv.ParseUint(router.Fields.CurrentStake, 10, 64)
// 		if err != nil {
// 			continue
// 		}
// 		network.Routers = append(network.Routers, &Router{
// 			XchgAddr:     router.Fields.XchgAddress,
// 			CurrentStake: stake,
// 			IpAddr:       router.Fields.IpAddr,
// 		})
// 	}
// 	return &network, nil
// }
