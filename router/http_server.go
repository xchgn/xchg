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
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type HttpServer struct {
	srv *http.Server
	//r                    *mux.Router
	server               *Router
	longPollingTimeout   time.Duration
	longPollingTickDelay time.Duration
	err                  error
}

func NewHttpServer() *HttpServer {
	var c HttpServer
	c.longPollingTimeout = 10 * time.Second
	c.longPollingTickDelay = 10 * time.Millisecond
	return &c
}

func (c *HttpServer) Start(server *Router, port int) {
	c.server = server
	c.srv = &http.Server{
		Addr: ":" + fmt.Sprint(port),
	}
	c.srv.Handler = c
	go c.thListen()
}

func (c *HttpServer) Stop() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = c.srv.Shutdown(ctx); err != nil {
		c.err = err
	}
	return err
}

func (c *HttpServer) thListen() {
	c.err = c.srv.ListenAndServe()
}

func (c *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.RequestURI == "/api/w" {
		c.processW(w, r)
		return
	}
	if r.RequestURI == "/api/r" {
		c.processR(w, r)
		return
	}
	if r.RequestURI == "/api/debug" {
		c.processDebug(w, r)
		return
	}
	if r.RequestURI == "/api/stat" {
		c.processStat(w, r)
		return
	}
	c.server.DeclareHttpRequestF()
	w.Write([]byte("wrong request"))
}

func (c *HttpServer) processR(w http.ResponseWriter, r *http.Request) {
	c.server.DeclareHttpRequestR()

	if r.Method == "POST" {
		if err := r.ParseMultipartForm(1000000); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
	}

	data64 := r.FormValue("d")
	var dataBS []byte
	var err error
	dataBS, err = base64.StdEncoding.DecodeString(data64)
	if err != nil {
		return
	}

	var resultBS []byte
	beginLongPollingDT := time.Now()
	for time.Since(beginLongPollingDT) < c.longPollingTimeout {
		var count int
		resultBS, count, err = c.server.GetMessages(dataBS)
		if count > 0 || err != nil {
			break
		}
		if errors.Is(r.Context().Err(), context.Canceled) {
			break
		}
		time.Sleep(c.longPollingTickDelay)
	}
	if err != nil {
		return
	}
	resultStr := base64.StdEncoding.EncodeToString(resultBS)

	result := []byte(resultStr)
	_, _ = w.Write([]byte(result))
}

func (c *HttpServer) processW(w http.ResponseWriter, r *http.Request) {
	c.server.DeclareHttpRequestW()

	if r.Method == "POST" {
		if err := r.ParseMultipartForm(1000000); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
	}

	data64 := r.FormValue("d")
	var dataBS []byte
	var err error
	dataBS, err = base64.StdEncoding.DecodeString(data64)
	if err != nil {
		return
	}

	offset := 0
	for offset < len(dataBS) {
		if offset+69 <= len(dataBS) {
			frameLen := int(binary.LittleEndian.Uint32(dataBS[offset:]))
			if offset+frameLen <= len(dataBS) {
				c.server.Put(dataBS[offset : offset+frameLen])
			} else {
				break
			}
			offset += frameLen
		} else {
			break
		}
	}
}

func (c *HttpServer) processDebug(w http.ResponseWriter, _ *http.Request) {
	c.server.DeclareHttpRequestD()
	result := []byte(c.server.DebugString())
	_, _ = w.Write(result)
}

func (c *HttpServer) processStat(w http.ResponseWriter, _ *http.Request) {
	c.server.DeclareHttpRequestS()
	result := []byte(c.server.StatString())
	_, _ = w.Write(result)
}
