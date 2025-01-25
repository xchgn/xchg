package blockchain

type FundSuiObj struct {
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

type RouterInfo struct {
	Fields struct {
		CurrentStake string `json:"currentStake"`
		IpAddr       string `json:"ipAddr"`
		XchgAddress  string `json:"xchgAddress"`
	} `json:"fields"`
	Type string `json:"type"`
}

type NetworkSegment struct {
	DataType string `json:"dataType"`
	Fields   struct {
		ID struct {
			ID string `json:"id"`
		} `json:"id"`
		Name  int `json:"name"`
		Value struct {
			Fields struct {
				Index   int          `json:"index"`
				Routers []RouterInfo `json:"routers"`
			} `json:"fields"`
			Type string `json:"type"`
		} `json:"value"`
	} `json:"fields"`
	HasPublicTransfer bool   `json:"hasPublicTransfer"`
	Type              string `json:"type"`
}
