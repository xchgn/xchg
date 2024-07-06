package xchg

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type Network struct {
	//mtx sync.Mutex

	Source string

	//fromInternet       bool
	//fromInternetLoaded bool

	Name          string   `json:"name"`
	Timestamp     int64    `json:"timestamp"`
	InitialPoints []string `json:"initial_points"`
	Ranges        []*rng   `json:"ranges"`
}

type host struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

func NewHost(address string) *host {
	var c host
	c.Address = address
	c.Name = "MainNet"
	return &c
}

type rng struct {
	Prefix string  `json:"prefix"`
	Hosts  []*host `json:"hosts"`
}

func NewRange(prefix string) *rng {
	var c rng
	c.Prefix = strings.ToLower(prefix)
	c.Hosts = make([]*host, 0)
	return &c
}

func NewNetwork() *Network {
	var c Network
	c.init()
	return &c
}

func NewNetworkFromBytes(rawContent []byte) (*Network, error) {
	var c Network
	c.init()
	json.Unmarshal(rawContent, &c)
	return &c, nil
}

func NewNetworkFromFile(fileName string) (*Network, error) {
	var bs []byte
	var err error

	bs, err = os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	network, err := NewNetworkFromBytes(bs)
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (c *Network) init() {
	c.Name = "MainNet"
	c.Timestamp = time.Now().Unix()

	c.Ranges = make([]*rng, 0)

	c.InitialPoints = make([]string, 0)
}

func (c *Network) GetRouterAddr() string {
	return "localhost:8084"
}
