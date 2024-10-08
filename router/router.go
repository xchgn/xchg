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

package router

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"
)

const (
	VERSION = int(24)
)

type Router struct {
	// Sync
	mtx sync.Mutex

	// State
	started  bool
	stopping bool

	// Data
	//nonces *Nonces

	//network *Network
	nextId uint64

	addresses map[string]*Storage

	// Statistics
	stat       RouterStatistics
	statLast   RouterStatistics
	statLastDT time.Time
	statSpeed  RouterSpeedStatistics

	lastDebugInfo []byte
	lastStatInfo  []byte

	httpServer *HttpServer

	clearAddressesLastDT time.Time
}

type RouterStatistics struct {
	FramesIn  int `json:"frames_in"`
	FramesOut int `json:"frames_out"`
	BytesIn   int `json:"bytes_in"`
	BytesOut  int `json:"bytes_out"`

	HttpRequests   int `json:"http_requests"`
	HttpRequestsR  int `json:"http_requests_r"`
	HttpRequestsW  int `json:"http_requests_w"`
	HttpRequestsN  int `json:"http_requests_n"`
	HttpRequestsNS int `json:"http_requests_ns"`
	HttpRequestsD  int `json:"http_requests_d"`
	HttpRequestsS  int `json:"http_requests_s"`
	HttpRequestsF  int `json:"http_requests_f"`
}

type RouterSpeedStatistics struct {
	SpeedHttpRequests   int `json:"http_requests"`
	SpeedHttpRequestsR  int `json:"http_requests_r"`
	SpeedHttpRequestsW  int `json:"http_requests_w"`
	SpeedHttpRequestsN  int `json:"http_requests_n"`
	SpeedHttpRequestsNS int `json:"http_requests_ns"`
	SpeedHttpRequestsD  int `json:"http_requests_d"`
	SpeedHttpRequestsF  int `json:"http_requests_f"`

	SpeedFramesIn  int `json:"frames_in"`
	SpeedFramesOut int `json:"frames_out"`
	SpeedBytesIn   int `json:"bytes_in"`
	SpeedBytesOut  int `json:"bytes_out"`

	SpeedBytesKIn  int `json:"kilobytes_in"`
	SpeedBytesKOut int `json:"kilobytes_out"`

	SpeedBytesMIn  int `json:"megabytes_in"`
	SpeedBytesMOut int `json:"megabytes_out"`

	Version int `json:"version"`
}

const (
	NONCE_COUNT       = 1024 * 1024
	INPUT_BUFFER_SIZE = 10 * 1024 * 1024
	STORING_TIMEOUT   = 60 * time.Second
)

func NewRouter() *Router {
	var c Router
	c.addresses = make(map[string]*Storage)
	c.nextId = 1

	c.statLastDT = time.Now()
	c.clearAddressesLastDT = time.Now()
	return &c
}

func (c *Router) Start() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Checks
	if c.started {
		return errors.New("already started")
	}
	if c.stopping {
		return errors.New("it is stopping")
	}

	go c.thBackgroundOperations()

	c.httpServer = NewHttpServer()
	c.httpServer.Start(c, 8084)

	return nil
}

func (c *Router) Stop() error {

	if c.httpServer != nil {
		c.httpServer.Stop()
		c.httpServer = nil
	}

	c.mtx.Lock()
	if !c.started {
		c.mtx.Unlock()
		return errors.New("already stopped")
	}
	if c.stopping {
		c.mtx.Unlock()
		return errors.New("already stopping")
	}
	c.stopping = true
	c.mtx.Unlock()

	for {
		c.mtx.Lock()
		started := c.started
		c.mtx.Unlock()
		if !started {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (c *Router) thBackgroundOperations() {
	c.started = true
	for {
		c.mtx.Lock()
		stopping := c.stopping
		c.mtx.Unlock()
		if stopping {
			break
		}
		time.Sleep(50 * time.Millisecond)
		c.thStatistics()
		c.thClearAddresses()
	}
	c.started = false
}

func (c *Router) thStatistics() {
	now := time.Now()
	if now.Sub(c.statLastDT) >= 1*time.Second {
		c.mtx.Lock()
		var stat RouterStatistics
		stat.BytesIn = c.stat.BytesIn - c.statLast.BytesIn
		stat.BytesOut = c.stat.BytesOut - c.statLast.BytesOut
		stat.FramesIn = c.stat.FramesIn - c.statLast.FramesIn
		stat.FramesOut = c.stat.FramesOut - c.statLast.FramesOut

		stat.HttpRequests = c.stat.HttpRequests - c.statLast.HttpRequests
		stat.HttpRequestsR = c.stat.HttpRequestsR - c.statLast.HttpRequestsR
		stat.HttpRequestsW = c.stat.HttpRequestsW - c.statLast.HttpRequestsW
		stat.HttpRequestsN = c.stat.HttpRequestsN - c.statLast.HttpRequestsN
		stat.HttpRequestsNS = c.stat.HttpRequestsNS - c.statLast.HttpRequestsNS
		stat.HttpRequestsD = c.stat.HttpRequestsD - c.statLast.HttpRequestsD
		stat.HttpRequestsF = c.stat.HttpRequestsF - c.statLast.HttpRequestsF

		c.statLast = c.stat
		c.mtx.Unlock()

		c.statSpeed.SpeedBytesIn = int(float64(stat.BytesIn) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedBytesOut = int(float64(stat.BytesOut) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedFramesIn = int(float64(stat.FramesIn) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedFramesOut = int(float64(stat.FramesOut) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedBytesKIn = c.statSpeed.SpeedBytesIn / 1024
		c.statSpeed.SpeedBytesKOut = c.statSpeed.SpeedBytesOut / 1024
		c.statSpeed.SpeedBytesMIn = c.statSpeed.SpeedBytesIn / (1024 * 1024)
		c.statSpeed.SpeedBytesMOut = c.statSpeed.SpeedBytesOut / (1024 * 1024)

		c.statSpeed.SpeedHttpRequests = int(float64(stat.HttpRequests) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsR = int(float64(stat.HttpRequestsR) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsW = int(float64(stat.HttpRequestsW) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsN = int(float64(stat.HttpRequestsN) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsNS = int(float64(stat.HttpRequestsNS) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsD = int(float64(stat.HttpRequestsD) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.SpeedHttpRequestsF = int(float64(stat.HttpRequestsF) / now.Sub(c.statLastDT).Seconds())
		c.statSpeed.Version = VERSION

		c.statLastDT = now
		c.buildDebugString()
	}
}

func (c *Router) thClearAddresses() {
	now := time.Now()
	if now.Sub(c.clearAddressesLastDT) >= 1*time.Second {
		c.mtx.Lock()
		addresses := make([]*Storage, 0)
		for address, addressStorage := range c.addresses {
			if now.Sub(addressStorage.TouchDT) > 10*time.Second {
				delete(c.addresses, address)
				continue
			}
			addresses = append(addresses, addressStorage)
		}
		c.mtx.Unlock()

		for _, a := range addresses {
			a.Clear()
		}

		c.clearAddressesLastDT = now
	}
}

func (c *Router) Put(frame []byte) {
	var ok bool
	var addressStorage *Storage

	addressDest := frame[64 : 64+32]
	// addressSrc := frame[32 : 32+32]

	// logger.Println("ROUTER PUT ", utils.TransactionSummary(frame))

	c.mtx.Lock()
	addrDestStr := hex.EncodeToString(addressDest)
	//addrSrcStr := hex.EncodeToString(addressSrc)
	//fmt.Println("ROUTER dest:", addrDestStr)
	//fmt.Println("ROUTER src:", addrSrcStr)
	addressStorage, ok = c.addresses[addrDestStr]
	if !ok || addressStorage == nil {
		addressStorage = NewStorage()
		c.addresses[addrDestStr] = addressStorage
	}
	id := c.nextId
	c.nextId++
	c.mtx.Unlock()

	addressStorage.Put(id, frame)
	//fmt.Println("ROUTER PUT:", tp, len(frame), id)

	c.stat.FramesIn++
	c.stat.BytesIn += len(frame)
}

// Get message request
func (c *Router) GetMessages(frame []byte) (response []byte, count int, err error) {
	var ok bool
	var addressStorage *Storage

	if len(frame) < 46 {
		err = errors.New("wrong frame size")
		return
	}

	afterId := binary.LittleEndian.Uint64(frame[0:])
	maxSize := binary.LittleEndian.Uint64(frame[8:])
	addressSrcBS := frame[16 : 16+32]

	addressSrc := addressSrcBS

	c.mtx.Lock()
	addressStorage, ok = c.addresses[hex.EncodeToString(addressSrc)]
	c.mtx.Unlock()

	if !ok || addressStorage == nil {
		response = make([]byte, 8)
		binary.LittleEndian.PutUint64(response[0:], 0)
		return
	}

	var msgData []byte
	var lastId uint64
	msgData, lastId, count = addressStorage.GetMessage(afterId, maxSize)
	response = make([]byte, 8+len(msgData))
	binary.LittleEndian.PutUint64(response[0:], lastId)
	if msgData != nil {
		copy(response[8:], msgData)
	}

	c.stat.FramesOut += count
	c.stat.BytesOut += len(msgData)
	return
}

func (c *Router) DebugString() (result []byte) {
	c.mtx.Lock()
	result = make([]byte, len(c.lastDebugInfo))
	copy(result, c.lastDebugInfo)
	c.mtx.Unlock()
	return
}

func (c *Router) StatString() (result []byte) {
	c.mtx.Lock()
	result = make([]byte, len(c.lastStatInfo))
	copy(result, c.lastStatInfo)
	c.mtx.Unlock()
	return
}

func (c *Router) DeclareHttpRequestR() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsR++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestW() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsW++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestN() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsN++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestNS() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsNS++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestD() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsD++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestS() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsS++
	c.mtx.Unlock()
}

func (c *Router) DeclareHttpRequestF() {
	c.mtx.Lock()
	c.stat.HttpRequests++
	c.stat.HttpRequestsF++
	c.mtx.Unlock()
}

func (c *Router) buildDebugString() {
	type AddressInfo struct {
		Address      string `json:"address"`
		MessageCount int    `json:"messages"`
	}

	type DebugInfo struct {
		AddressCount int                   `json:"address_count"`
		NextMsgId    int                   `json:"next_msg_id"`
		Stat         RouterStatistics      `json:"stat_total"`
		StatSpeed    RouterSpeedStatistics `json:"stat_in_second"`
		Addresses    []AddressInfo         `json:"addresses"`
	}

	c.mtx.Lock()
	var di DebugInfo
	di.AddressCount = len(c.addresses)
	di.NextMsgId = int(c.nextId)
	di.Stat = c.stat
	di.StatSpeed = c.statSpeed

	di.Addresses = make([]AddressInfo, 0, len(c.addresses))
	for address, a := range c.addresses {
		var ai AddressInfo
		ai.Address = address
		ai.MessageCount = a.MessagesCount()
		di.Addresses = append(di.Addresses, ai)
	}
	c.mtx.Unlock()

	sort.Slice(di.Addresses, func(i, j int) bool {
		return di.Addresses[i].Address < di.Addresses[j].Address
	})

	bsDebug, _ := json.MarshalIndent(di, "", " ")
	bsStatSpeed, _ := json.MarshalIndent(di.StatSpeed, "", " ")
	c.mtx.Lock()
	c.lastDebugInfo = bsDebug
	c.lastStatInfo = bsStatSpeed
	c.mtx.Unlock()

	// logger.Println("stat", string(bsJson))
}

func CheckHash(hash []byte, complexity byte) bool {
	if len(hash) != 32 {
		return false
	}
	mask := make([]byte, (complexity/8)+1)
	for w := 0; w < len(mask); w++ {
		mask[w] = 0x00
	}

	for q := byte(0); q < complexity; q++ {
		byteIndex := int(q / 8)
		bitIndex := q % 8
		mask[byteIndex] = mask[byteIndex] | (0x80 >> bitIndex)
	}

	for k := 0; k < len(mask); k++ {
		if hash[k]&mask[k] != 0 {
			return false
		}
	}

	return true
}
