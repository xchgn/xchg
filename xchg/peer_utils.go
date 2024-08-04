package xchg

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
)

func (c *Peer) httpCall(httpClient *http.Client, routerHost string, function string, frame []byte) (result []byte, err error) {
	if len(routerHost) == 0 {
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

	response, err := c.Post(httpClient, addr+"/api/"+function, writer.FormDataContentType(), &body, "https://"+addr)

	if err != nil {
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

func (c *Peer) Post(httpClient *http.Client, url, contentType string, body io.Reader, host string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}

func (c *Peer) send(frame []byte) {
	addr := c.network.GetRouterAddr()
	go c.httpCall(c.httpClient, addr, "w", frame)
}

func (c *Peer) fixStat() {
	c.mtx.Lock()
	c.logger.Println("-------STAT------")
	for key, value := range c.routerStatRead {
		c.logger.Println("Router read", key, "=", value)
	}
	c.logger.Println()
	c.mtx.Unlock()
}
