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
	"sync"
	"time"
)

type Storage struct {
	mtx         sync.Mutex
	TouchDT     time.Time
	maxMessages int
	messages    []*Message
}

func NewStorage() *Storage {
	var c Storage
	c.maxMessages = 100000000
	c.messages = make([]*Message, 0, c.maxMessages+1)
	c.TouchDT = time.Now()
	return &c
}

func (c *Storage) Clear() {
	now := time.Now()
	c.mtx.Lock()
	oldMessages := c.messages
	c.messages = make([]*Message, 0, len(oldMessages))
	for _, m := range oldMessages {
		if now.Sub(m.TouchDT) < 5*time.Second {
			c.messages = append(c.messages, m)
		}
	}
	c.mtx.Unlock()
}

func (c *Storage) MessagesCount() (count int) {
	c.mtx.Lock()
	count = len(c.messages)
	c.mtx.Unlock()
	return
}

func (c *Storage) Put(id uint64, frame []byte) {
	c.mtx.Lock()
	msg := NewMessage(id, frame)
	c.messages = append(c.messages, msg)
	if len(c.messages) > c.maxMessages {
		c.messages = c.messages[1:]
	}
	c.TouchDT = time.Now()
	c.mtx.Unlock()
}

func (c *Storage) GetMessage(afterId uint64, maxSize uint64) (data []byte, lastId uint64, count int) {
	//fmt.Println("addressStorage.GetMessage", afterId)

	data = make([]byte, 0)
	lastId = afterId
	count = 0
	sendAll := false
	c.mtx.Lock()

	if len(c.messages) > 0 {
		if afterId > c.messages[len(c.messages)-1].id {
			afterId = c.messages[0].id
			sendAll = true
		}
	}

	for _, m := range c.messages {
		if m.id > afterId || sendAll {
			if len(data)+len(m.data) < int(maxSize) {
				data = append(data, m.data...)
				lastId = m.id
				count++
			}
		}
	}

	if len(data) == 0 && len(c.messages) > 0 {
		if afterId > c.messages[len(c.messages)-1].id {
			lastId = c.messages[0].id
		}
	}
	c.mtx.Unlock()
	return
}
