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
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

type Server struct {
	serverConnection *xchg.Peer
	privateKey       *ecdsa.PrivateKey
	accessKey        string
	processor        func(param *xchg.Param) (response []byte, err error)
}

func StartServer(privateKey *ecdsa.PrivateKey, accessKey string, processor func(param *xchg.Param) (response []byte, err error)) *Server {
	var c Server
	c.privateKey = privateKey
	c.accessKey = accessKey
	c.serverConnection = xchg.NewPeer(privateKey)
	c.serverConnection.Callback = c.ServerProcessorCall
	c.processor = processor
	c.serverConnection.Start()
	return &c
}

func StartServerFast(accessKey string, processor func(param *xchg.Param) (response []byte, err error)) *Server {
	privateKey, _ := utils.GeneratePrivateKey()
	s := StartServer(privateKey, accessKey, processor)
	s.privateKey = privateKey
	return s
}

func (c *Server) Address() common.Address {
	return utils.PublicKeyToAddress(&c.privateKey.PublicKey)
}

func (c *Server) Stop() {
	c.serverConnection.Stop()
}

func (c *Server) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == c.accessKey {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *Server) ServerProcessorCall(param *xchg.Param) (response []byte, err error) {
	if c.processor != nil {

		return c.processor(param)
	}
	return nil, errors.New("not implemented")
}
