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
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/xchgn/xchg/router"
	"github.com/xchgn/xchg/utils"
)

type Logger interface {
	Println(v ...interface{})
}

type Param struct {
	LocalPeer     *Peer
	RemoteAddress ed25519.PublicKey
	AuthData      []byte
	Function      string
	Parameter     []byte
}

type CallbackFunc func(param *Param) (response []byte, err error)

type Peer struct {
	mtx        sync.Mutex
	privateKey ed25519.PrivateKey
	started    bool
	stopping   bool
	network    *Network

	TransportPrivateKey []byte
	TransportPublicKey  []byte

	httpClient     *http.Client
	httpClientLong *http.Client

	longPollingDelay time.Duration

	localAddressBS []byte

	logger         Logger
	routerStatRead map[string]int

	gettingFromInternet   map[string]bool
	lastReceivedMessageId map[string]uint64

	// Client
	remotePeers map[string]*RemotePeer

	// Server
	incomingTransactions map[string]*Transaction
	sessionsById         map[uint64]*Session
	authNonces           *Nonces
	nextSessionId        uint64

	Callback CallbackFunc

	lastPurgeSessionsTime time.Time

	router1 *router.Router
}

type Session struct {
	id              uint64
	aesKey          []byte
	authData        []byte
	remotePublicKey ed25519.PublicKey
	lastAccessDT    time.Time
	snakeCounter    *SnakeCounter
}

func NewPeer(privateKey ed25519.PrivateKey) *Peer {
	var c Peer
	c.logger = NewDefaultLogger()
	c.remotePeers = make(map[string]*RemotePeer)
	c.incomingTransactions = make(map[string]*Transaction)
	c.authNonces = NewNonces(100)
	c.sessionsById = make(map[uint64]*Session)
	c.nextSessionId = 1
	c.network = NewNetwork()
	c.lastReceivedMessageId = make(map[string]uint64)

	c.routerStatRead = make(map[string]int)

	c.TransportPrivateKey, c.TransportPublicKey, _ = utils.GenerateCurve25519KeyPair()

	c.gettingFromInternet = make(map[string]bool)
	c.longPollingDelay = 12 * time.Second

	c.privateKey = privateKey
	if c.privateKey == nil {
		c.privateKey, _ = utils.GeneratePrivateKey()
	}

	{
		tr := &http.Transport{}
		jar, _ := cookiejar.New(nil)
		c.httpClient = &http.Client{Transport: tr, Jar: jar}
		c.httpClient.Timeout = 2 * time.Second
	}

	{
		tr := &http.Transport{}
		jar, _ := cookiejar.New(nil)
		c.httpClientLong = &http.Client{Transport: tr, Jar: jar}
		c.httpClientLong.Timeout = c.longPollingDelay
	}

	return &c
}

func StartServerPeer(privateKey ed25519.PrivateKey, callback CallbackFunc) *Peer {
	c := NewPeer(privateKey)
	c.Callback = callback
	c.Start()
	return c
}

func StartClientPeer() *Peer {
	c := NewPeer(nil)
	c.Start()
	return c
}

func (c *Peer) Start() (err error) {
	c.logger.Println("Peer::Start")
	c.mtx.Lock()
	if c.started {
		c.mtx.Unlock()
		err = errors.New("already started")
		return
	}
	c.mtx.Unlock()

	c.localAddressBS = utils.ExtractPublicKey(c.privateKey)

	c.router1 = router.NewRouter()
	c.router1.Start()

	go c.thWork()

	return
}

func (c *Peer) Stop() (err error) {
	fmt.Println("Peer::Stop")

	c.mtx.Lock()
	if !c.started {
		c.mtx.Unlock()
		err = errors.New("already stopped")
		return
	}

	c.stopping = true
	started := c.started
	c.mtx.Unlock()

	if c.router1 != nil {
		c.router1.Stop()
		c.router1 = nil
	}

	dtBegin := time.Now()
	for started {
		time.Sleep(10 * time.Millisecond)
		c.mtx.Lock()
		started = c.started
		c.mtx.Unlock()

		if time.Since(dtBegin) > 1000*time.Millisecond {
			fmt.Println("TIMEOUT")
			break
		}
	}

	c.mtx.Lock()
	started = c.started
	c.mtx.Unlock()

	if started {
		err = errors.New("timeout")
	}

	return
}

func (c *Peer) Address() ed25519.PublicKey {
	return utils.ExtractPublicKey(c.privateKey)
}

func (c *Peer) Network() *Network {
	return c.network
}

func (c *Peer) thWork() {
	c.started = true
	lastPurgeSessionsDT := time.Now()
	lastStatDT := time.Now()
	for {
		c.mtx.Lock()
		stopping := c.stopping
		c.mtx.Unlock()
		if stopping {
			break
		}

		go c.getFramesFromRouters()

		if time.Since(lastPurgeSessionsDT) > 5*time.Second {
			c.purgeSessions()
			lastPurgeSessionsDT = time.Now()
		}

		if time.Since(lastStatDT) > 10*time.Second {
			c.fixStat()
			lastStatDT = time.Now()
		}

		time.Sleep(10 * time.Millisecond)
	}
	c.started = false
}

func (c *Peer) Call(remoteAddress ed25519.PublicKey, authData string, function string, data []byte, timeout time.Duration) (result []byte, err error) {
	c.mtx.Lock()
	remotePeer, remotePeerOk := c.remotePeers[hex.EncodeToString(remoteAddress)]
	if !remotePeerOk || remotePeer == nil {
		remotePeer = NewRemotePeer(remoteAddress, authData, c.privateKey)
		c.remotePeers[hex.EncodeToString(remoteAddress)] = remotePeer
	}
	network := c.network
	c.mtx.Unlock()
	result, err = remotePeer.Call(network, function, data, timeout)
	return
}
