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
	"os"
	"path/filepath"
	"strings"
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

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
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

func (c *HttpServer) thListen() {
	c.err = c.srv.ListenAndServe()
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

func (c *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	c.processFile(w, r)
}

func (c *HttpServer) processDebug(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	c.server.DeclareHttpRequestD()
	result := []byte(c.server.DebugString())
	_, _ = w.Write(result)
}

func (c *HttpServer) processStat(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	c.server.DeclareHttpRequestS()
	result := []byte(c.server.StatString())
	_, _ = w.Write(result)
}

func (c *HttpServer) processR(w http.ResponseWriter, r *http.Request) {
	c.server.DeclareHttpRequestR()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

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
	// fmt.Println("ROUTER processR", dataBS)

	//addrTemp := base32.StdEncoding.EncodeToString(dataBS[16 : 16+30])

	//addrTemp = strings.ToLower(addrTemp)

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
	/*if err != nil {
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}*/
	//fmt.Println("ROUTER read result:", len(result))
	_, _ = w.Write([]byte(result))
}

func (c *HttpServer) processW(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("processW")
	c.server.DeclareHttpRequestW()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

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
				//fmt.Println("Router Write")
				c.server.Put(dataBS[offset : offset+frameLen])
				/*if err != nil {
					return
				}*/
			} else {
				break
			}
			offset += frameLen
		} else {
			break
		}
	}

	/*if err != nil {
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}*/
}

func SplitRequest(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	})
}

func (c *HttpServer) processFile(w http.ResponseWriter, _ *http.Request) {
	c.server.DeclareHttpRequestF()
	w.Write([]byte("wrong request"))
}

/*func getRealAddr(r *http.Request) string {

	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP

}

func (c *HttpServer) redirect(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Accel-Expires", "0")
	http.Redirect(w, r, url, 307)
}
*/
