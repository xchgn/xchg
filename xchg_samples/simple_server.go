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
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/xchgn/xchg/xchg"
)

type SimpleServer struct {
	serverConnection *xchg.Peer
	defaultResponse  []byte
}

func NewSimpleServer(privateKey *ecdsa.PrivateKey) *SimpleServer {
	var c SimpleServer
	c.serverConnection = xchg.NewPeer(privateKey, xchg.NewDefaultLogger())
	c.serverConnection.ServerProcessorAuth = c.ServerProcessorAuth
	c.serverConnection.ServerProcessorCall = c.ServerProcessorCall

	c.defaultResponse = make([]byte, 10)
	rand.Read(c.defaultResponse)
	return &c
}

func (c *SimpleServer) Start() {
	c.serverConnection.Start(true)
}

func (c *SimpleServer) Stop() {
	c.serverConnection.Stop()
}

func (c *SimpleServer) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == "pass" {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *SimpleServer) ServerProcessorCall(authData []byte, function string, parameter []byte) (response []byte, err error) {
	switch function {
	case "version":
		//response = []byte("simple server 2.42 0123456789|0123456789|0123456789|0123456789")
		//response = make([]byte, 3000)
		//rand.Read(response)
		response = []byte(fmt.Sprint(time.Now().Unix()))
	case "time":
		strTime := time.Now().String()
		response = []byte(strTime)
	case "json-api":
		type InputStruct struct {
			A int
			B int
		}
		var request InputStruct
		err = json.Unmarshal(parameter, &request)
		if err != nil {
			return
		}
		type OutputStruct struct {
			C int
		}
		var resp OutputStruct
		resp.C = request.A + request.B
		response, err = json.MarshalIndent(resp, "", " ")
	default:
		err = errors.New(xchg.ERR_XCHG_NOT_IMPLEMENTED)
	}

	return
}
