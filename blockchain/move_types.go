package blockchain

type MoveFundSuiObj struct {
	DataType string `json:"dataType"`
	Fields   struct {
		Addresses  MoveTable     `json:"addresses"`
		Balance    string        `json:"balance"`
		CommonFund string        `json:"commonFund"`
		Counter    string        `json:"counter"`
		Directors  []interface{} `json:"directors"`
		ID         struct {
			ID string `json:"id"`
		} `json:"id"`
		Network    MoveTable `json:"network"`
		Parameters struct {
			Fields struct {
				Contents []interface{} `json:"contents"`
			} `json:"fields"`
			Type string `json:"type"`
		} `json:"parameters"`
		Profiles                    MoveTable     `json:"profiles"`
		ProposalsToChangeParameters []interface{} `json:"proposalsToChangeParameters"`
		ProposalsToSpend            []interface{} `json:"proposalsToSpend"`

		Routers MoveTable `json:"routers"`
	} `json:"fields"`
	HasPublicTransfer bool   `json:"hasPublicTransfer"`
	Type              string `json:"type"`
}

type MoveTable struct {
	Fields struct {
		ID struct {
			ID string `json:"id"`
		} `json:"id"`
		Size string `json:"size"`
	} `json:"fields"`
	Type string `json:"type"`
}

type MoveRouterInfo struct {
	Fields struct {
		CurrentStake string `json:"currentStake"`
		IpAddr       string `json:"ipAddr"`
		XchgAddress  string `json:"xchgAddress"`
	} `json:"fields"`
	Type string `json:"type"`
}

type MoveNetworkSegment struct {
	DataType string `json:"dataType"`
	Fields   struct {
		ID struct {
			ID string `json:"id"`
		} `json:"id"`
		Name  int `json:"name"`
		Value struct {
			Fields struct {
				Index   int              `json:"index"`
				Routers []MoveRouterInfo `json:"routers"`
			} `json:"fields"`
			Type string `json:"type"`
		} `json:"value"`
	} `json:"fields"`
	HasPublicTransfer bool   `json:"hasPublicTransfer"`
	Type              string `json:"type"`
}

/*
Router Object
{
"dataType": "moveObject",
  "fields": {
    "id": {
      "id": "0x185c9202168190fd7ab541eb1fa8d2bf158f23a1368be6ce1d5b3f5aaf9c66cd"
    },
    "name": "0x9337d82d8b18a3fdf294d020319483e3f383716fbe89490955ff71b4cb518a76",
    "value": {
      "fields": {
        "chequeIds": {
          "fields": {
            "contents": [
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e0",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e1",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e2",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e3",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e4",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e5",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e6",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e7",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e8",
              "0x00000000000000000000000000000000000000000000000000062c9b1c3f00e9"
            ]
          },
          "type": "0x2::vec_set::VecSet\u003caddress\u003e"
        },
        "ipAddr": "123.123.123.123",
        "name": "qwe",
        "owner": "0x8c1bd6aea293d2a425ae89a793cdff6bf6163c9b0c541af8b0f2609a230513fe",
        "segment": 3,
        "totalStakeAmount": "42000000"
      },
      "type": "0x7674735315216af4fc71ef20bb5775ad3b1e1f162c19150f0da9772f142528dd::fund::Router"
    }
  },
  "hasPublicTransfer": false,
  "type": "0x2::dynamic_field::Field\u003caddress, 0x7674735315216af4fc71ef20bb5775ad3b1e1f162c19150f0da9772f142528dd::fund::Router\u003e"
}*/

type MoveRouterObject struct {
	DataType string `json:"dataType"`
	Fields   struct {
		ID struct {
			ID string `json:"id"`
		} `json:"id"`
		Name  string `json:"name"`
		Value struct {
			Fields struct {
				ChequeIds struct {
					Fields struct {
						Contents []string `json:"contents"`
					} `json:"fields"`
					Type string `json:"type"`
				} `json:"chequeIds"`
				IpAddr         string `json:"ipAddr"`
				Name           string `json:"name"`
				Owner          string `json:"owner"`
				Segment        int    `json:"segment"`
				TotalStakeAmnt string `json:"totalStakeAmount"`
			} `json:"fields"`
			Type string `json:"type"`
		} `json:"value"`
	} `json:"fields"`
	HasPublicTransfer bool   `json:"hasPublicTransfer"`
	Type              string `json:"type"`
}
