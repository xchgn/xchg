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
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xchgn/xchg/utils"
)

type RemotePeerTransport interface {
	Id() string
	Check(frame20 *Transaction, network *Network, remotePublicKeyExists bool) error
	DeclareError(sentViaTransportMap map[string]struct{})
	Send(network *Network, tr *Transaction) error
	SetRemoteUDPAddress(udpAddress *net.UDPAddr)
}

type RemotePeer struct {
	mtx           sync.Mutex
	remoteAddress common.Address
	authData      string
	//network       *Network

	privateKey      *ecdsa.PrivateKey
	remotePublicKey *ecdsa.PublicKey

	nonces *Nonces

	httpClient *http.Client

	//findingConnection    bool
	authProcessing       bool
	aesKey               []byte
	sessionId            uint64
	sessionNonceCounter  uint64
	outgoingTransactions map[uint64]*Transaction
	nextTransactionId    uint64
}

func NewRemotePeer(remoteAddress common.Address, authData string, privateKey *ecdsa.PrivateKey) *RemotePeer {
	var c RemotePeer
	c.privateKey = privateKey
	c.remoteAddress = remoteAddress
	c.authData = authData
	c.outgoingTransactions = make(map[uint64]*Transaction)
	c.nextTransactionId = 1
	c.nonces = NewNonces(100)

	tr := &http.Transport{}
	jar, _ := cookiejar.New(nil)
	c.httpClient = &http.Client{Transport: tr, Jar: jar}
	c.httpClient.Timeout = 1 * time.Second

	return &c
}

func (c *RemotePeer) RemoteAddress() common.Address {
	return c.remoteAddress
}

func (c *RemotePeer) processFrame(routerHost string, frame []byte) {
	frameType := frame[4]

	switch frameType {
	case FrameTypeResponse:
		c.processFrame11(routerHost, frame)
	}
}

func (c *RemotePeer) processFrame11(routerHost string, frame []byte) {
	_ = routerHost
	transaction, err := Parse(frame)
	if err != nil {
		fmt.Println(err)
		return
	}

	if transaction.TotalSize > XchgMaxTransactionSize {
		fmt.Println("Error: transaction.TotalSize > XchgMaxTransactionSize")
		return
	}

	c.mtx.Lock()
	if t, ok := c.outgoingTransactions[transaction.TransactionId]; ok {
		if transaction.Err == nil {
			t.AppendReceivedData(transaction)
		} else {
			t.Result = transaction.Data
			t.Err = transaction.Err
			t.Complete = true
		}
	}
	c.mtx.Unlock()
}

func (c *RemotePeer) setConnectionPoint(routerHost string, publicKey *ecdsa.PublicKey) {
	_ = routerHost
	c.mtx.Lock()
	c.remotePublicKey = publicKey
	c.mtx.Unlock()
	//fmt.Println("Received Address for", c.remoteAddress, "from", routerHost)
}

func (c *RemotePeer) Call(network *Network, function string, data []byte, timeout time.Duration) (result []byte, err error) {
	c.mtx.Lock()
	sessionId := c.sessionId
	c.mtx.Unlock()

	if c.remotePublicKey == nil {
		addressBS := c.remoteAddress.Bytes()
		generatedLocalCheque := &Cheque{}

		transaction := NewTransaction(XchgFrameGetPublicKeyRequest,
			utils.PublicKeyToAddress(&c.privateKey.PublicKey),
			c.remoteAddress,
			0,
			0,
			0,
			0,
			generatedLocalCheque,
			nil)
		transaction.Data = make([]byte, XchgAddressSize)

		copy(transaction.Data, addressBS)
		addr := network.GetRouterAddr()
		c.httpCall(addr, "w", transaction.Marshal())

		// Wait for public key for 1 second
		for i := 0; i < 100; i++ {
			time.Sleep(10 * time.Millisecond)
			if c.remotePublicKey != nil {
				break
			}
		}
	}

	if c.remotePublicKey == nil {
		return nil, errors.New("NO PUBLIC KEY")
	}

	if sessionId == 0 {
		err = c.auth(network, 1000*time.Millisecond)
		if err != nil {
			return
		}
	}

	c.mtx.Lock()
	aesKey := make([]byte, len(c.aesKey))
	copy(aesKey, c.aesKey)
	c.mtx.Unlock()

	result, err = c.regularCall(network, function, data, aesKey, timeout)

	return
}

func (c *RemotePeer) auth(network *Network, timeout time.Duration) (err error) {
	//fmt.Println("RemotePeer::auth")

	c.mtx.Lock()
	if c.authProcessing {
		c.mtx.Unlock()
		err = errors.New("auth processing")
		return
	}
	c.authProcessing = true
	c.mtx.Unlock()

	defer func() {
		c.mtx.Lock()
		c.authProcessing = false
		c.mtx.Unlock()
	}()

	var nonce []byte
	nonce, err = c.regularCall(network, "/xchg-get-nonce", nil, nil, timeout)
	if err != nil {
		fmt.Println("get nonce error", err)
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_GET_NONCE + ":" + err.Error())
		return
	}
	if len(nonce) != 16 {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_WRONG_NONCE_LEN)
		return
	}

	c.mtx.Lock()
	localPrivateKey := c.privateKey
	remotePublicKey := c.remotePublicKey
	authData := make([]byte, len(c.authData))
	copy(authData, []byte(c.authData))
	c.mtx.Unlock()

	if c.privateKey == nil {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_NO_LOCAL_PRIVATE_KEY)
		return
	}

	if remotePublicKey == nil {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_NO_REMOTE_PUBLIC_KEY)
		return
	}

	localPublicKeyBS := utils.PublicKeyToBytes(&localPrivateKey.PublicKey)

	authFrameSecret := make([]byte, 16+len(authData))
	copy(authFrameSecret[0:], nonce)
	copy(authFrameSecret[16:], authData)
	var encryptedAuthFrame []byte
	encryptedAuthFrame, err = utils.EncryptBytesWithPublicKey(remotePublicKey, []byte(authFrameSecret))
	if err != nil {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_ENC + ":" + err.Error())
		return
	}

	authFrame := make([]byte, len(localPublicKeyBS)+len(encryptedAuthFrame))
	copy(authFrame, localPublicKeyBS)
	copy(authFrame[len(localPublicKeyBS):], encryptedAuthFrame)

	var result []byte
	result, err = c.regularCall(network, "/xchg-auth", authFrame, nil, timeout)
	if err != nil {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_AUTH + ":" + err.Error())
		return
	}

	result, err = utils.DecryptBytesWithPrivateKey(localPrivateKey, result)
	if err != nil {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_DECR + ":" + err.Error())
		return
	}

	if len(result) != 8+32 {
		err = errors.New(ERR_XCHG_CL_CONN_AUTH_WRONG_AUTH_RESP_LEN)
		return
	}

	c.mtx.Lock()
	c.sessionId = binary.LittleEndian.Uint64(result)
	c.aesKey = make([]byte, 32)
	copy(c.aesKey, result[8:])
	c.mtx.Unlock()

	return
}

func (c *RemotePeer) regularCall(network *Network, function string, data []byte, aesKey []byte, timeout time.Duration) (result []byte, err error) {
	//fmt.Println("RemotePeer::regularCall", function)

	if len(function) > 255 {
		err = errors.New(ERR_XCHG_CL_CONN_CALL_WRONG_FUNCTION_LEN)
		return
	}

	encryptedWithAES := false

	c.mtx.Lock()
	localPrivateKey := c.privateKey
	c.mtx.Unlock()

	if localPrivateKey == nil {
		err = errors.New(ERR_XCHG_CL_CONN_CALL_NO_LOCAL_PRIVATE_KEY)
		return
	}

	c.mtx.Lock()
	sessionId := c.sessionId
	c.mtx.Unlock()

	c.mtx.Lock()
	sessionNonceCounter := c.sessionNonceCounter
	c.sessionNonceCounter++
	c.mtx.Unlock()

	var frame []byte
	if len(aesKey) == 32 {
		// session is active - using AES
		// [0:8] = sessionNonceCounter
		// [8] = len(function)
		// [9:n] = function
		// [n:m] = data
		frame = make([]byte, 8+1+len(function)+len(data))
		binary.LittleEndian.PutUint64(frame, sessionNonceCounter)
		frame[8] = byte(len(function))
		copy(frame[9:], function)
		copy(frame[9+len(function):], data)
		frame = utils.PackBytes(frame)
		frame, err = utils.EncryptAESGCM(frame, aesKey)
		if err != nil {
			c.Reset()
			err = errors.New(ERR_XCHG_CL_CONN_CALL_ENC + ":" + err.Error())
			return
		}
		encryptedWithAES = true
	} else {
		// session is not active - using ECDSA
		// [0] = len(function)
		// [1:n] = function
		// [n:m] = data
		frame = make([]byte, 1+len(function)+len(data))
		frame[0] = byte(len(function))
		copy(frame[1:], function)
		copy(frame[1+len(function):], data)
	}

	result, err = c.executeTransaction(network, sessionId, frame, timeout, aesKey)

	if NeedToChangeNode(err) {
		c.Reset()
		return
	}

	if err != nil {
		err = errors.New(ERR_XCHG_CL_CONN_CALL_ERR + ":" + err.Error())
		return
	}

	fmt.Println("Received result:", len(result))

	if encryptedWithAES {
		result, err = utils.DecryptAESGCM(result, aesKey)
		if err != nil {
			c.Reset()
			err = errors.New(ERR_XCHG_CL_CONN_CALL_DECRYPT + ":" + err.Error())
			return
		}
		result, err = utils.UnpackBytes(result)
		if err != nil {
			c.Reset()
			err = errors.New(ERR_XCHG_CL_CONN_CALL_UNPACK + ":" + err.Error())
			return
		}
	}

	if len(result) < 1 {
		err = errors.New(ERR_XCHG_CL_CONN_CALL_RESP_LEN)
		c.Reset()
		return
	}

	if result[0] == 0 {
		// Success response
		result = result[1:]
		err = nil
		return
	}

	if result[0] == 1 {
		// Error response
		err = errors.New(ERR_XCHG_CL_CONN_CALL_FROM_PEER + ":" + string(result[1:]))
		if NeedToMakeSession(err) {
			// Any server error - make new session
			c.sessionId = 0
		}
		result = nil
		return
	}

	err = errors.New(ERR_XCHG_CL_CONN_CALL_RESP_STATUS_BYTE)
	c.Reset()
	return
}

func (c *RemotePeer) Reset() {
	c.mtx.Lock()
	c.reset()
	c.mtx.Unlock()
}

func (c *RemotePeer) reset() {
	c.sessionId = 0
	c.aesKey = nil
}

func (c *RemotePeer) executeTransaction(network *Network, sessionId uint64, data []byte, timeout time.Duration, aesKeyOriginal []byte) (result []byte, err error) {

	// Get transaction ID
	var transactionId uint64
	c.mtx.Lock()
	transactionId = c.nextTransactionId
	c.nextTransactionId++

	generatedLocalCheque := &Cheque{}

	// Create transaction
	t := NewTransaction(FrameTypeCall, utils.PublicKeyToAddress(&c.privateKey.PublicKey), c.remoteAddress, transactionId, sessionId, 0, len(data), generatedLocalCheque, data)
	c.outgoingTransactions[transactionId] = t
	c.mtx.Unlock()

	// Send transaction
	sentCount := 0
	sendCounter := 0
	offset := 0
	for offset < len(data) {
		sendCounter++
		currentBlockSize := XchgMaxFrameSize
		restDataLen := len(data) - offset
		if restDataLen < currentBlockSize {
			currentBlockSize = restDataLen
		}

		blockTransaction := NewTransaction(FrameTypeCall, utils.PublicKeyToAddress(&c.privateKey.PublicKey), c.remoteAddress, transactionId, sessionId, offset, len(data), generatedLocalCheque, data[offset:offset+currentBlockSize])

		err = c.Send(network, blockTransaction)

		if err != nil {
			c.mtx.Lock()
			delete(c.outgoingTransactions, t.TransactionId)
			c.mtx.Unlock()
			return
		}
		sentCount++
		offset += currentBlockSize
	}

	if sendCounter != sentCount {
		return nil, errors.New("no route")
	}

	// Wait for response
	waitingDurationInMilliseconds := timeout.Milliseconds()
	waitingTick := int64(10)
	waitingIterationCount := waitingDurationInMilliseconds / waitingTick
	for i := int64(0); i < waitingIterationCount; i++ {
		if t.Complete {
			// Transaction complete
			c.mtx.Lock()
			delete(c.outgoingTransactions, t.TransactionId)
			c.mtx.Unlock()

			// Error recevied
			if t.Err != nil {
				result = nil
				err = t.Err
				return
			}

			// Success
			result = t.Result
			err = nil
			return
		}
		time.Sleep(time.Duration(waitingTick) * time.Millisecond)
	}

	// Clear transactions map
	c.mtx.Lock()
	delete(c.outgoingTransactions, t.TransactionId)
	c.mtx.Unlock()

	fmt.Println("exec transaction timeout")

	c.mtx.Lock()

	allowResetSession := true
	if len(aesKeyOriginal) == 32 && len(c.aesKey) == 32 {
		for i := 0; i < 32; i++ {
			if aesKeyOriginal[i] != c.aesKey[i] {
				allowResetSession = false
				break
			}
		}
	}

	if allowResetSession {
		c.sessionId = 0
		c.aesKey = nil
	}

	c.mtx.Unlock()

	return nil, errors.New(ERR_XCHG_PEER_CONN_TR_TIMEOUT)
}

func (c *RemotePeer) httpCall(routerHost string, function string, frame []byte) (result []byte, err error) {
	if len(routerHost) == 0 {
		return
	}

	if len(frame) < TransactionHeaderSize {
		return
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	{
		fw, _ := writer.CreateFormField("d")
		frame64 := base64.StdEncoding.EncodeToString(frame)
		fw.Write([]byte(frame64))
	}
	writer.Close()

	addr := "http://" + routerHost
	uri := addr + "/api/" + function

	response, err := c.Post(uri, writer.FormDataContentType(), &body, addr)

	if err != nil {
		fmt.Println("HTTP error:", err)
		return
	} else {
		var content []byte
		content, err = io.ReadAll(response.Body)
		if err != nil {
			response.Body.Close()
			return
		}
		result, err = base64.StdEncoding.DecodeString(string(content))
		response.Body.Close()
	}
	return
}

func (c *RemotePeer) Post(url, contentType string, body io.Reader, host string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.httpClient.Do(req)
}

func (c *RemotePeer) Send(network *Network, tr *Transaction) (err error) {
	addr := network.GetRouterAddr()
	bs := tr.Marshal()
	c.httpCall(addr, "w", bs)
	return
}
