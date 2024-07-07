// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Xchg-Network Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
