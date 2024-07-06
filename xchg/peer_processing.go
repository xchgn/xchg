package xchg

import (
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
)

func (c *Peer) processFrame(routerHost string, frame []byte) (responseFrames []*Transaction) {
	if len(frame) < TransactionHeaderSize {
		return
	}

	frameType := frame[4]

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
	var processor ServerProcessor

	_ = routerHost

	responseFrames = make([]*Transaction, 0)

	transaction, err := Parse(frame)
	if err != nil {
		return
	}

	c.mtx.Lock()
	processor = c.processor
	var incomingTransaction *Transaction

	for trCode, tr := range c.incomingTransactions {
		now := time.Now()
		if now.Sub(tr.BeginDT) > 10*time.Second {
			delete(c.incomingTransactions, trCode)
		}
	}

	srcAddr, _ := utils.BytesToAddress(transaction.SrcAddress[:])

	var ok bool
	incomingTransactionCode := fmt.Sprint(transaction.SrcAddress, "-", transaction.TransactionId)
	if incomingTransaction, ok = c.incomingTransactions[incomingTransactionCode]; !ok {
		incomingTransaction = NewTransaction(transaction.FrameType, utils.PublicKeyToAddress(&c.privateKey.PublicKey), srcAddr, transaction.TransactionId, transaction.SessionId, 0, int(transaction.TotalSize), make([]byte, 0))
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

	srcAddress, _ := utils.BytesToAddress(transaction.SrcAddress[:])

	if processor != nil {
		resp, dontSendResponse := c.onEdgeReceivedCall(incomingTransaction.SessionId, incomingTransaction.Data)
		if !dontSendResponse {
			trResponse := NewTransaction(0x11, utils.PublicKeyToAddress(&c.privateKey.PublicKey), srcAddress, incomingTransaction.TransactionId, incomingTransaction.SessionId, 0, len(resp), resp)

			offset := 0
			blockSize := XchgMaxFrameSize
			for offset < len(trResponse.Data) {
				currentBlockSize := blockSize
				restDataLen := len(trResponse.Data) - offset
				if restDataLen < currentBlockSize {
					currentBlockSize = restDataLen
				}

				blockTransaction := NewTransaction(0x11, utils.PublicKeyToAddress(&c.privateKey.PublicKey), srcAddress, trResponse.TransactionId, trResponse.SessionId, offset, len(resp), trResponse.Data[offset:offset+currentBlockSize])
				blockTransaction.Offset = uint32(offset)
				blockTransaction.TotalSize = uint32(len(trResponse.Data))
				blockTransaction.FromLocalNode = incomingTransaction.FromLocalNode
				responseFrames = append(responseFrames, blockTransaction)
				offset += currentBlockSize
			}
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

		srcAddress, _ := utils.BytesToAddress(tr.SrcAddress[:])
		if peer.RemoteAddress().Hex() == srcAddress.Hex() {
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
	if len(transaction.Data) != XchgAddressSize {
		return
	}
	if len(c.localAddressBS) != XchgAddressSize {
		return
	}

	// Compare requested address and local address
	requestedAddress := transaction.Data[:XchgAddressSize]
	for i := 0; i < 20; i++ {
		if c.localAddressBS[i] != requestedAddress[i] {
			return // It is not my address
		}
	}

	// Send Public Key
	publicKeyBS := utils.PublicKeyToBytes(&c.privateKey.PublicKey)
	srcAddress, _ := utils.BytesToAddress(transaction.SrcAddress[:])
	response := NewTransaction(XchgFrameGetPublicKeyResponse, utils.PublicKeyToAddress(&c.privateKey.PublicKey), srcAddress, 0, 0, 0, 0, nil)
	response.Data = make([]byte, XchgPublicKeySize)
	copy(response.Data, publicKeyBS)
	responseFrames = append(responseFrames, response)
	return
}

func (c *Peer) processFrameGetPublicKeyResponse(routerHost string, frame []byte) {
	transaction, err := Parse(frame)
	if err != nil {
		return
	}
	if len(transaction.Data) != XchgPublicKeySize {
		return
	}
	receivedPublicKeyBS := transaction.Data
	receivedPublicKey, err := utils.BytesToPublicKey([]byte(receivedPublicKeyBS))
	if err != nil {
		return
	}
	receivedAddress := utils.PublicKeyToAddress(receivedPublicKey)
	c.mtx.Lock()

	for _, peer := range c.remotePeers {
		if peer.RemoteAddress().Hex() == receivedAddress.Hex() {
			peer.setConnectionPoint(routerHost, receivedPublicKey)
			break
		}
	}

	c.mtx.Unlock()
}
