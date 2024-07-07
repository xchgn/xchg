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
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xchgn/xchg/utils"
)

type Transaction struct {
	// Transport Header - 5 bytes
	Length    uint32 // 0
	FrameType byte   // 4

	// Call header - 24 bytes
	TransactionId uint64 // 5
	SessionId     uint64 // 13
	Offset        uint32 // 21
	TotalSize     uint32 // 25

	// 20 bytes - Source Address
	SrcAddress [utils.AddressBytesSize]byte // 29
	// 20 bytes - Source Address
	DestAddress [utils.AddressBytesSize]byte // 49

	//
	Cheque Cheque

	// Data
	Data []byte

	FromLocalNode bool

	// Execution Result
	BeginDT        time.Time
	ReceivedFrames []*Transaction
	//ReceivedDataLen int
	Complete bool
	Result   []byte
	Err      error
}

const (
	TransactionHeaderSize = 69

	FrameTypeCall     = byte(0x10)
	FrameTypeResponse = byte(0x11)
)

func NewTransaction(frameType byte, srcAddress common.Address, destAddress common.Address, transactionId uint64, sessionId uint64, offset int, totalSize int, data []byte) *Transaction {
	var c Transaction
	c.Length = uint32(TransactionHeaderSize + len(data))
	c.FrameType = frameType

	c.TransactionId = transactionId
	c.SessionId = sessionId
	c.Offset = uint32(offset)
	c.TotalSize = uint32(totalSize)

	copy(c.SrcAddress[:], srcAddress.Bytes())
	copy(c.DestAddress[:], destAddress.Bytes())

	c.Data = data

	c.ReceivedFrames = make([]*Transaction, 0)
	return &c
}

func (c *Transaction) SrcAddressString() string {
	address, _ := utils.BytesToAddress(c.SrcAddress[:])
	return address.Hex()
}

func (c *Transaction) DestAddressString() string {
	address, _ := utils.BytesToAddress(c.DestAddress[:])
	return address.Hex()
}

func Parse(frame []byte) (tr *Transaction, err error) {
	if len(frame) < TransactionHeaderSize {
		err = errors.New("wrong frame")
		return
	}

	tr = &Transaction{}
	tr.Length = binary.LittleEndian.Uint32(frame[0:])
	tr.FrameType = frame[4]

	tr.TransactionId = binary.LittleEndian.Uint64(frame[5:])
	tr.SessionId = binary.LittleEndian.Uint64(frame[13:])
	tr.Offset = binary.LittleEndian.Uint32(frame[21:])
	tr.TotalSize = binary.LittleEndian.Uint32(frame[25:])

	copy(tr.SrcAddress[:], frame[29:])
	copy(tr.DestAddress[:], frame[49:])

	tr.Data = make([]byte, len(frame)-TransactionHeaderSize)
	copy(tr.Data, frame[TransactionHeaderSize:])

	return
}

func (c *Transaction) Marshal() (result []byte) {
	result = make([]byte, TransactionHeaderSize+len(c.Data))

	binary.LittleEndian.PutUint32(result[0:], uint32(len(result))) // Length
	result[4] = c.FrameType

	binary.LittleEndian.PutUint64(result[5:], c.TransactionId)
	binary.LittleEndian.PutUint64(result[13:], c.SessionId)
	binary.LittleEndian.PutUint32(result[21:], c.Offset)
	binary.LittleEndian.PutUint32(result[25:], c.TotalSize)

	copy(result[29:], c.SrcAddress[:])
	copy(result[49:], c.DestAddress[:])

	copy(result[TransactionHeaderSize:], c.Data)

	return
}

func (c *Transaction) String() string {
	res := fmt.Sprint(c.TransactionId) + "t:" + fmt.Sprint(c.FrameType) + " dl:" + fmt.Sprint(len(c.Data))
	return res
}

func (c *Transaction) AppendReceivedData(transaction *Transaction) {
	if len(c.ReceivedFrames) < 1000 {
		found := false
		for _, trvTr := range c.ReceivedFrames {
			if trvTr.Offset == transaction.Offset {
				found = true
			}
		}
		if !found {
			c.ReceivedFrames = append(c.ReceivedFrames, transaction)
		}
	}

	if transaction.FromLocalNode {
		c.FromLocalNode = true
	}

	receivedDataLen := 0
	for _, trvTr := range c.ReceivedFrames {
		receivedDataLen += int(len(trvTr.Data))
	}

	if receivedDataLen == int(transaction.TotalSize) {
		if len(c.Result) != int(transaction.TotalSize) {
			c.Result = make([]byte, transaction.TotalSize)
		}
		for _, trvTr := range c.ReceivedFrames {
			copy(c.Result[trvTr.Offset:], trvTr.Data)
		}
		c.Complete = true
	}
}
