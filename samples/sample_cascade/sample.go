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

package samplecascade

import (
	"fmt"
	"time"

	"github.com/ipoluianov/gomisc/logger"
	"github.com/xchgn/xchg/xchg"
)

/*
	peer1->call("A");
		b = peer2->call("B")
		c = peer2->call("C")
	return b+c
*/

func Run() {
	peer1 := xchg.StartServerPeer(nil, func(param *xchg.Param) (response []byte, err error) {
		if param.Function == "" {
			return
		}
		if param.Function == "get_name_and_status" {
			// logger.Println("nested calling:", hex.EncodeToString(param.RemoteAddress))
			responseName, err := param.LocalPeer.Call(param.RemoteAddress, "", "get_name", nil, 2*time.Second)
			if err != nil {
				fmt.Println("peer1 error:", err)
				return nil, err
			}

			responseStatus, err := param.LocalPeer.Call(param.RemoteAddress, "", "get_status", nil, time.Second)
			if err != nil {
				fmt.Println("peer1 error:", err)
				return nil, err
			}

			response = []byte(string(responseName) + ":" + string(responseStatus))
			//response = []byte("qqq")

		}
		return
	})
	peer2 := xchg.StartServerPeer(nil, func(param *xchg.Param) (response []byte, err error) {
		if param.Function == "" {
			return
		}
		if param.Function == "get_name" {
			response = []byte("MyName")
			return
		}
		if param.Function == "get_status" {
			response = []byte("MyStatus")
			return
		}
		return
	})

	logger.Println("peer1", peer1.AddressHex())
	logger.Println("peer2", peer2.AddressHex())

	response, err := peer2.Call(peer1.Address(), "", "get_name_and_status", nil, 3*time.Second)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(string(response))
	}
}
