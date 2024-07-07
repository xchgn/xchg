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
	"crypto/rand"
	"encoding/binary"
	"sync"
)

type Nonces struct {
	mtx          sync.Mutex
	nonces       [][16]byte
	currentIndex int
	complexity   byte
}

func NewNonces(size int) *Nonces {
	var c Nonces
	c.complexity = 0
	c.nonces = make([][16]byte, size)
	for i := 0; i < size; i++ {
		c.fillNonce(i)
	}
	c.currentIndex = 0
	return &c
}

func (c *Nonces) fillNonce(index int) {
	if index >= 0 && index < len(c.nonces) {
		binary.LittleEndian.PutUint32(c.nonces[index][:], uint32(index)) // Index of nonce for search
		c.nonces[index][4] = c.complexity                                // Current Complexity
		rand.Read(c.nonces[index][5:])                                   // Random Nonce
	}
}

func (c *Nonces) Next() [16]byte {
	var result [16]byte
	c.mtx.Lock()
	c.fillNonce(c.currentIndex)
	result = c.nonces[c.currentIndex]
	c.currentIndex++
	if c.currentIndex >= len(c.nonces) {
		c.currentIndex = 0
	}
	c.mtx.Unlock()
	return result
}

func (c *Nonces) Check(nonce []byte) bool {
	result := true
	c.mtx.Lock()
	index := int(binary.LittleEndian.Uint32(nonce[:]))
	if index >= 0 && index < len(c.nonces) {
		for i := 0; i < 16; i++ {
			if c.nonces[index][i] != nonce[i] {
				result = false
				break
			}
		}
	} else {
		result = false
	}
	if result {
		c.fillNonce(index)
	}
	c.mtx.Unlock()
	return result
}
