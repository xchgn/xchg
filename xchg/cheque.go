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

/*
import (
	"encoding/binary"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xchgn/xchg/utils"
)

type Cheque struct {
	mtx         sync.Mutex
	DT          int64                   // 8
	DestAddress common.Address          // 20
	Value       int64                   // 8
	Signature   [XchgSignatureSize]byte // 65
}

func (c *Cheque) SerializeToBytes() []byte {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	result := make([]byte, XchgChequeSize)
	binary.LittleEndian.PutUint64(result[0:], uint64(c.DT))
	copy(result[8:], c.DestAddress.Bytes())
	binary.LittleEndian.PutUint64(result[8+XchgAddressSize:], uint64(c.DT))
	copy(result[XchgChequeDataSize:], c.Signature[:])
	return result
}

func (c *Cheque) DeserializeFromBytes(data []byte) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if len(data) != XchgChequeSize {
		return
	}
	c.DT = int64(binary.LittleEndian.Uint64(data[0:]))
	c.DestAddress.SetBytes(data[8 : 8+XchgAddressSize])
	c.Value = int64(binary.LittleEndian.Uint64(data[8+XchgAddressSize:]))
	copy(c.Signature[:], data[8+XchgAddressSize+8:])
}

func (c *Cheque) Check() (checkResult bool, address common.Address) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	chequeAsBytes := c.SerializeToBytes()
	publicKey, err := utils.RecoverPublicKeyFromSignature(chequeAsBytes[:XchgChequeDataSize], c.Signature[:])
	if err != nil {
		checkResult = false
		return
	}
	address = crypto.PubkeyToAddress(*publicKey)
	checkResult = utils.VerifySignature(publicKey, chequeAsBytes[:XchgChequeDataSize], c.Signature[:])
	return
}
*/
