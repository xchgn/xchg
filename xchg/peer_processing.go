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
	"time"

	"github.com/xchgn/xchg/utils"
)

func (c *Peer) processFrame(routerHost string, frame []byte) (responseFrames []*Transaction) {
	if len(frame) < TransactionHeaderSize {
		return
	}

	frameType := frame[4]

	//fmt.Println("processFrame", frameType)

	switch frameType {
	case XchgFrameCallRequest:
		responseFrames = c.processFrameCallRequest(routerHost, frame)
	case XchgFrameCallResponse:
		c.processFrameCallResponse(routerHost, frame)
	case XchgFrameGetPublicKeyRequest:
		responseFrames = c.processFrameGetPublicKeyRequest(frame)
	case XchgFrameGetPublicKeyResponse:
		c.processFrameGetPublicKeyResponse(routerHost, frame)
	}

	return
}

func (c *Peer) processFrameCallRequest(routerHost string, frame []byte) (responseFrames []*Transaction) {

	_ = routerHost

	responseFrames = make([]*Transaction, 0)

	transaction, err := Parse(frame)
	if err != nil {
		return
	}

	c.mtx.Lock()
	var incomingTransaction *Transaction

	for trCode, tr := range c.incomingTransactions {
		now := time.Now()
		if now.Sub(tr.BeginDT) > 10*time.Second {
			delete(c.incomingTransactions, trCode)
		}
	}

	srcAddr := transaction.SrcAddress[:]

	publicKey := utils.ExtractPublicKey(c.privateKey)

	var ok bool
	incomingTransactionCode := fmt.Sprint(transaction.SrcAddress, "-", transaction.TransactionId)
	if incomingTransaction, ok = c.incomingTransactions[incomingTransactionCode]; !ok {
		incomingTransaction = NewTransaction(transaction.FrameType,
			srcAddr,
			publicKey,
			transaction.TransactionId,
			transaction.SessionId,
			0,
			int(transaction.TotalSize),
			make([]byte, 0))
		incomingTransaction.BeginDT = time.Now()
		c.incomingTransactions[incomingTransactionCode] = incomingTransaction
	}

	incomingTransaction.AppendReceivedData(transaction)

	if incomingTransaction.Complete {
		incomingTransaction.Data = incomingTransaction.Result
		incomingTransaction.Result = nil
	} else {
		c.mtx.Unlock()
		return
	}

	delete(c.incomingTransactions, incomingTransactionCode)
	c.mtx.Unlock()

	srcAddress := transaction.SrcAddress[:]

	//generatedLocalCheque := &Cheque{}

	fmt.Println("---", hex.EncodeToString(incomingTransaction.SrcAddress[:]))

	resp, dontSendResponse := c.onEdgeReceivedCall(incomingTransaction.SessionId, incomingTransaction.Data, incomingTransaction.SrcAddress[:])
	if !dontSendResponse {
		trResponse := NewTransaction(0x11,
			publicKey,
			srcAddress,
			incomingTransaction.TransactionId,
			incomingTransaction.SessionId,
			0,
			len(resp),
			resp)

		offset := 0
		blockSize := XchgMaxFrameSize
		for offset < len(trResponse.Data) {
			currentBlockSize := blockSize
			restDataLen := len(trResponse.Data) - offset
			if restDataLen < currentBlockSize {
				currentBlockSize = restDataLen
			}

			blockTransaction := NewTransaction(0x11,
				publicKey,
				srcAddress,
				trResponse.TransactionId,
				trResponse.SessionId,
				offset,
				len(resp),

				trResponse.Data[offset:offset+currentBlockSize])

			blockTransaction.Offset = uint32(offset)
			blockTransaction.TotalSize = uint32(len(trResponse.Data))
			blockTransaction.FromLocalNode = incomingTransaction.FromLocalNode
			responseFrames = append(responseFrames, blockTransaction)
			offset += currentBlockSize
		}
	}
	return
}

func (c *Peer) processFrameCallResponse(routerHost string, frame []byte) {
	tr, err := Parse(frame)
	if err != nil {
		return
	}

	var remotePeer *RemotePeer
	// Find the peer in local remote peers collection
	c.mtx.Lock()
	for _, peer := range c.remotePeers {

		srcAddress := tr.SrcAddress[:]
		if hex.EncodeToString(peer.RemoteAddress()) == hex.EncodeToString(srcAddress) {
			remotePeer = peer
			break
		}
	}
	c.mtx.Unlock()
	if remotePeer != nil {
		remotePeer.processFrame(routerHost, frame)
	}
}

// Get Public Key Request
// This frame received by server
// It converts the address to the public key
func (c *Peer) processFrameGetPublicKeyRequest(frame []byte) (responseFrames []*Transaction) {
	responseFrames = make([]*Transaction, 0)
	transaction, err := Parse(frame)

	// Checks
	if err != nil {
		return
	}

	if len(c.localAddressBS) != XchgAddressSize {
		return
	}

	// Send Public Key
	response := NewTransaction(XchgFrameGetPublicKeyResponse,
		c.localAddressBS,
		transaction.SrcAddress[:],
		0,
		0,
		0,
		0,
		nil)
	response.Data = make([]byte, 32+64)
	copy(response.Data, c.TransportPublicKey)
	signature := utils.SignMessage(c.privateKey, c.TransportPublicKey)
	copy(response.Data[32:], signature)

	responseFrames = append(responseFrames, response)
	return
}

func (c *Peer) processFrameGetPublicKeyResponse(routerHost string, frame []byte) {
	transaction, err := Parse(frame)
	if err != nil {
		return
	}
	if len(transaction.Data) != 32+64 {
		return
	}
	remoteTransportPublicKey := transaction.Data[:32]
	signature := transaction.Data[32:]

	verifyResult := utils.VerifySignature(transaction.SrcAddress[:], remoteTransportPublicKey, signature)
	if !verifyResult {
		return
	}

	c.mtx.Lock()
	for _, peer := range c.remotePeers {
		if hex.EncodeToString(peer.RemoteAddress()) == transaction.SrcAddressString() {
			peer.setRemoteTransportPublicKey(routerHost, remoteTransportPublicKey)
			break
		}
	}
	c.mtx.Unlock()
}
