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

package xchg_samples

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xchgn/xchg/xchg"
)

type SimpleClient struct {
	address common.Address
	client  *xchg.Peer
}

func NewSimpleClient(address common.Address) *SimpleClient {
	var c SimpleClient
	c.address = address
	c.client = xchg.NewPeer(nil)
	c.client.Start()
	return &c
}

func (c *SimpleClient) Version() (result string, err error) {
	var resultBS []byte
	resultBS, err = c.client.Call(c.address, "pass", "version", nil, 2*time.Second)
	result = string(resultBS)
	return
}

func (c *SimpleClient) Calculate(a int, b int) (result int, err error) {
	c.client.Call(c.address, "pass", "static-string", nil, time.Second)
	return
}
