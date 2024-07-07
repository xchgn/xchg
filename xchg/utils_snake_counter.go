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
	"errors"
	"fmt"
	"sync"
)

type SnakeCounter struct {
	mtx           sync.Mutex
	size          int
	data          []byte
	lastProcessed int
}

func NewSnakeCounter(size int, initValue int) *SnakeCounter {
	var c SnakeCounter
	c.size = size
	c.lastProcessed = -1
	c.data = make([]byte, size)
	for i := 0; i < c.size; i++ {
		c.data[i] = 1
	}
	c.TestAndDeclare(initValue)
	return &c
}

func (c *SnakeCounter) TestAndDeclare(counter int) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if counter < c.lastProcessed-len(c.data) {
		return errors.New("too less")
	}

	if counter > c.lastProcessed {
		shiftRange := counter - c.lastProcessed
		newData := make([]byte, c.size)
		for i := 0; i < len(c.data); i++ {
			b := byte(0)
			oldAddressOfCell := i - shiftRange
			if oldAddressOfCell >= 0 && oldAddressOfCell < len(c.data) {
				b = c.data[oldAddressOfCell]
			}
			newData[i] = b
		}
		c.data = newData
		c.data[0] = 1
		c.lastProcessed = counter
		return nil
	}

	index := c.lastProcessed - counter
	if index >= 0 && index < c.size {

		if c.data[index] == 0 {
			c.data[c.lastProcessed-counter] = 1
			return nil
		}
	}

	return errors.New("already used")
}

func (c *SnakeCounter) LastProcessed() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return c.lastProcessed
}

func (c *SnakeCounter) Print() {
	fmt.Println("--------------------")
	fmt.Println("Header:", c.lastProcessed)
	for i, v := range c.data {
		fmt.Println(c.lastProcessed-i, v)
	}
	fmt.Println("--------------------")
}
