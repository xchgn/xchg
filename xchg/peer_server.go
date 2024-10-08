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
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"log"
	"time"

	"github.com/xchgn/xchg/utils"
)

const (
	INTERNAL_ERROR = "#internal_error#"
)

func (c *Peer) onEdgeReceivedCall(sessionId uint64, data []byte, remoteRealPublicKey ed25519.PublicKey) (response []byte, dontSendResponse bool) {

	var err error
	// Find the session
	var session *Session
	encryped := false
	if sessionId != 0 {
		c.mtx.Lock()
		var ok bool
		session, ok = c.sessionsById[sessionId]
		if !ok {
			session = nil
		}
		c.mtx.Unlock()
	}

	if sessionId != 0 {
		if session == nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_SESSION))
			return
		}
		data, err = utils.DecryptAESGCM(data, session.aesKey)
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_DECR + ":" + err.Error()))
			return
		}
		data, err = utils.Unpack(data)
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_UNPACK + ":" + err.Error()))
			return
		}
		if len(data) < 9 {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN9))
			return
		}
		encryped = true
		callNonce := binary.LittleEndian.Uint64(data)
		err = session.snakeCounter.TestAndDeclare(int(callNonce))
		if err != nil {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_NONCE))
			dontSendResponse = true
			return
		}
		data = data[8:]
		session.lastAccessDT = time.Now()
	} else {
		if len(data) < 1 {
			response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN1))
			return
		}
	}

	functionLen := int(data[0])
	if len(data) < 1+functionLen {
		response = prepareResponseError(errors.New(ERR_XCHG_SRV_CONN_WRONG_LEN_FN))
		return
	}
	function := string(data[1 : 1+functionLen])
	functionParameter := data[1+functionLen:]

	var callFunc CallbackFunc
	c.mtx.Lock()
	callFunc = c.Callback
	c.mtx.Unlock()

	var resp []byte

	if sessionId == 0 {
		switch function {
		case "/xchg-get-nonce":
			nonce := c.authNonces.Next()
			resp = nonce[:]
		case "/xchg-auth":
			resp, err = c.processAuth(functionParameter, remoteRealPublicKey)
			if err != nil {
				if err.Error() == INTERNAL_ERROR {
					dontSendResponse = true
					return
				}
				return
			}
		}
	} else {
		authData := make([]byte, 0)
		if session != nil {
			authData = session.authData
		}
		var p Param
		p.AuthData = authData
		p.Function = function
		p.Parameter = functionParameter
		p.LocalPeer = c
		p.RemoteAddress = session.remoteRealPublicKey
		resp, err = callFunc(&p)
	}

	if err != nil {
		response = prepareResponseError(err)
	} else {
		response = prepareResponse(resp)
	}

	if encryped {
		response = utils.Pack(response)
		response, err = utils.EncryptAESGCM(response, session.aesKey)
		if err != nil {
			return
		}
	}

	return
}

func (c *Peer) processAuth(functionParameter []byte, remoteRealPublicKey ed25519.PublicKey) (response []byte, err error) {
	if len(functionParameter) < XchgPublicKeySize {
		err = errors.New(INTERNAL_ERROR)
		return
	}

	remoteTransportPublicKeyBS := functionParameter[:XchgPublicKeySize]
	aesKey, _ := utils.GetSharedKey(c.TransportPrivateKey, remoteTransportPublicKeyBS)

	encryptedAuthFrame := functionParameter[XchgPublicKeySize:]
	parameter, err := utils.DecryptAESGCM(encryptedAuthFrame, aesKey)
	if err != nil {
		err = errors.New(INTERNAL_ERROR)
		return
	}

	if len(parameter) < XchgNonceSize {
		err = errors.New(INTERNAL_ERROR)
		return
	}

	nonce := parameter[0:XchgNonceSize]
	if !c.authNonces.Check(nonce) {
		err = errors.New(INTERNAL_ERROR)
		return
	}

	authData := parameter[XchgNonceSize:]

	callbackFunc := c.Callback

	var p Param
	p.LocalPeer = c
	p.RemoteAddress = remoteRealPublicKey
	p.AuthData = authData
	_, err = callbackFunc(&p)
	if err != nil {
		return
	}

	c.mtx.Lock()

	sessionId := c.nextSessionId
	c.nextSessionId++
	session := &Session{}
	session.id = sessionId
	session.lastAccessDT = time.Now()
	session.aesKey = aesKey
	session.snakeCounter = NewSnakeCounter(100, 0)
	session.authData = authData
	session.remoteRealPublicKey = remoteRealPublicKey
	c.sessionsById[sessionId] = session
	response = make([]byte, 8)
	binary.LittleEndian.PutUint64(response, sessionId)

	response, err = utils.EncryptAESGCM(response, aesKey)

	c.mtx.Unlock()

	return
}

func (c *Peer) purgeSessions() {
	c.logger.Println("Peer::purgeSessions")

	now := time.Now()
	c.mtx.Lock()
	if now.Sub(c.lastPurgeSessionsTime).Seconds() > 60 {
		for sessionId, session := range c.sessionsById {
			if now.Sub(session.lastAccessDT).Seconds() > 60 {
				delete(c.sessionsById, sessionId)
				log.Println("Session removed", sessionId)
			}
		}
		c.lastPurgeSessionsTime = time.Now()
	}
	c.mtx.Unlock()
}

func prepareResponseError(err error) []byte {
	errBS := make([]byte, 0)
	if err != nil {
		errBS = []byte(err.Error())
	}
	respFrame := make([]byte, 1+len(errBS))
	respFrame[0] = 1
	copy(respFrame[1:], errBS)
	return respFrame
}

func prepareResponse(data []byte) []byte {
	respFrame := make([]byte, 1+len(data))
	respFrame[0] = 0
	copy(respFrame[1:], data)
	return respFrame
}
