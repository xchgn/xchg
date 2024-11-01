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
	"encoding/hex"
	"fmt"

	"github.com/xchgn/xchg/utils"
)

type Network struct {
	//routers []*RouterConnection
}

type RouterInfo struct {
	Name        string
	NetAddress  string
	XchgAddress string
}

func NewNetwork() *Network {
	var c Network
	c.init()
	return &c
}

func (c *Network) init() {
}

func (c *Network) GetRouterAddr(address string) string {
	fmt.Println("GetRouterAddr for ", address)
	routers := c.GetRouters()
	for _, router := range routers {
		fmt.Println("Router:", router.Name, router.NetAddress, router.XchgAddress)
	}
	return "localhost:8084"
}

func (c *Network) GetRouters() []*RouterInfo {
	result := make([]*RouterInfo, 0)

	for i := 0; i < 10; i++ {
		privateKey, _ := utils.GeneratePrivateKey()
		publicKey := utils.ExtractPublicKey(privateKey)
		xchgAddress := hex.EncodeToString(publicKey)

		result = append(result, &RouterInfo{
			Name:        "router" + fmt.Sprint(i),
			NetAddress:  "localhost:8084",
			XchgAddress: xchgAddress,
		})
	}

	return result
}
