package xchg

import (
	"errors"
	"fmt"
	"time"

	"github.com/xchgn/xchg/utils"
)

func (c *Peer) processFrame(routerHost string, frame []byte) (responseFrames []*Transaction) {
	if len(frame) < TransactionHeaderSize {
		return
	}

	frameType := frame[4]

	// Call Request
	if frameType == 0x10 {
		responseFrames = c.processFrame10(routerHost, frame)
		return
	}

	// Call Response
	if frameType == 0x11 {
		c.processFrame11(routerHost, frame)
		return
	}

	// ARP request
	if frameType == 0x20 {
		responseFrames = c.processFrame20(frame)
		return
	}

	// ARP response
	if frameType == 0x21 {
		c.processFrame21(routerHost, frame)
		return
	}

	return
}

// ----------------------------------------
// Incoming Call - Server Role
// ----------------------------------------

func (c *Peer) processFrame10(routerHost string, frame []byte) (responseFrames []*Transaction) {
	var processor ServerProcessor

	responseFrames = make([]*Transaction, 0)

	transaction, err := Parse(frame)
	if err != nil {
		return
	}

	if c.network != nil {
		if c.network.IsLocalNode(routerHost) {
			transaction.FromLocalNode = true
		}
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
		incomingTransaction = NewTransaction(transaction.FrameType, utils.AddressForPublicKey(&c.privateKey.PublicKey), srcAddr, transaction.TransactionId, transaction.SessionId, 0, int(transaction.TotalSize), make([]byte, 0))
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
			trResponse := NewTransaction(0x11, utils.AddressForPublicKey(&c.privateKey.PublicKey), srcAddress, incomingTransaction.TransactionId, incomingTransaction.SessionId, 0, len(resp), resp)

			offset := 0
			blockSize := 4 * 1024
			for offset < len(trResponse.Data) {
				currentBlockSize := blockSize
				restDataLen := len(trResponse.Data) - offset
				if restDataLen < currentBlockSize {
					currentBlockSize = restDataLen
				}

				blockTransaction := NewTransaction(0x11, utils.AddressForPublicKey(&c.privateKey.PublicKey), srcAddress, trResponse.TransactionId, trResponse.SessionId, offset, len(resp), trResponse.Data[offset:offset+currentBlockSize])
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

func (c *Peer) processFrame11(routerHost string, frame []byte) {
	tr, err := Parse(frame)
	if err != nil {
		return
	}

	var remotePeer *RemotePeer
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

// ARP LAN request
func (c *Peer) processFrame20(frame []byte) (responseFrames []*Transaction) {

	responseFrames = make([]*Transaction, 0)

	//c.mtx.Lock()
	//localAddress := AddressForPublicKey(&c.privateKey.PublicKey)
	//c.mtx.Unlock()

	transaction, err := Parse(frame)
	if err != nil {
		return
	}

	fmt.Println("20 received", transaction)

	nonce := transaction.Data[:16]

	requestedAddress := transaction.Data[16:]

	for i := 0; i < len(requestedAddress); i++ { // TODO
		if c.localAddressBS[i] != requestedAddress[i] {
			return
		}
	}

	// Send my public key
	publicKeyBS := utils.PublicKeyToDer(&c.privateKey.PublicKey)

	// And signature
	signature, err := utils.SignData(c.privateKey, nonce)
	if err != nil {
		return
	}

	srcAddress, _ := utils.BytesToAddress(transaction.SrcAddress[:])

	response := NewTransaction(0x21, utils.AddressForPublicKey(&c.privateKey.PublicKey), srcAddress, 0, 0, 0, 0, nil)
	response.Data = make([]byte, 16+256+len(publicKeyBS))
	copy(response.Data[0:], nonce)
	copy(response.Data[16:], signature)
	copy(response.Data[16+256:], publicKeyBS)
	responseFrames = append(responseFrames, response)
	return
	//_, _ = conn.WriteTo(response.Marshal(), sourceAddress)
}

func (c *Peer) processFrame21(routerHost string, frame []byte) {
	transaction, err := Parse(frame)
	if err != nil {
		return
	}

	if len(transaction.Data) < 16+256 {
		err = errors.New("wrong frame size")
		return
	}

	receivedPublicKeyBS := transaction.Data[16+256:]
	receivedPublicKey, err := utils.PublicKeyFromDer([]byte(receivedPublicKeyBS))
	if err != nil {
		return
	}

	receivedAddress := utils.AddressForPublicKey(receivedPublicKey)

	c.mtx.Lock()

	for _, peer := range c.remotePeers {
		if peer.RemoteAddress() == receivedAddress {
			peer.setConnectionPoint(routerHost, receivedPublicKey, transaction.Data[0:16], transaction.Data[16:16+65])
			break
		}
	}

	c.mtx.Unlock()
}
