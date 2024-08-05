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

package validator

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type RouterInfo struct {
	Address             string `json:"Address"`
	HttpConnectionPoint string `json:"HttpConnectionPoint"`
	Signature           string `json:"signature"`
}

type Validator struct {
	privateKey *ecdsa.PrivateKey
	mtx        sync.Mutex
	routers    []*RouterInfo
	//routersMap map[string]*RouterInfo

	srv *http.Server
	err error
}

func NewValidator() *Validator {
	var c Validator

	return &c
}

func (c *Validator) Start() {
	c.routers = make([]*RouterInfo, 0)
	//c.privateKey, _ = utils.GeneratePrivateKey()
	c.srv = &http.Server{
		Addr: ":8184",
	}

	c.srv.Handler = c
	go c.thListen()
}

func (c *Validator) Stop() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = c.srv.Shutdown(ctx); err != nil {
		c.err = err
	}
	return err
}

func (c *Validator) thListen() {
	c.err = c.srv.ListenAndServe()
}

func (c *Validator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/api/routers" {
		c.processRouters(w, r)
		return
	}
	if r.RequestURI == "/api/register_router" {
		c.processRegisterRouter(w, r)
		return
	}
	w.Write([]byte("wrong request"))
}

func (c *Validator) processRouters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	c.mtx.Lock()
	routerAsBytes, err := json.MarshalIndent(c.routers, "", " ")
	c.mtx.Unlock()

	if err != nil {
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}

	w.Write(routerAsBytes)
}

func (c *Validator) processRegisterRouter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method == "POST" {
		if err := r.ParseMultipartForm(1000000); err != nil {
			return
		}
	}

	//addressStr := r.FormValue("address")
	ipAddress := r.FormValue("ipAddress")
	signature := r.FormValue("signature")

	//address, err := utils.StringToAddress(addressStr)
	/*if err != nil {
		return
	}*/

	var router RouterInfo
	//router.Address = address.Hex()
	router.HttpConnectionPoint = ipAddress
	router.Signature = signature
	c.routers = append(c.routers, &router)

	c.declareRouter(&router)

	/*if err != nil {
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}*/
}

func (c *Validator) declareRouter(routerInfo *RouterInfo) {
	_ = routerInfo
	c.mtx.Lock()
	defer c.mtx.Unlock()
}
