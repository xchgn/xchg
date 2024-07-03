package xchg_samples

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xchgn/xchg/utils"
	"github.com/xchgn/xchg/xchg"
)

type Server struct {
	serverConnection *xchg.Peer
	privateKey       *ecdsa.PrivateKey
	accessKey        string
	processor        func(function string, parameter []byte) (response []byte, err error)
}

func StartServer(privateKey *ecdsa.PrivateKey, accessKey string, processor func(function string, parameter []byte) (response []byte, err error)) *Server {
	var c Server
	c.privateKey = privateKey
	c.accessKey = accessKey
	c.serverConnection = xchg.NewPeer(privateKey, xchg.NewDefaultLogger())
	c.serverConnection.SetProcessor(&c)
	c.processor = processor
	c.serverConnection.Start(true)
	return &c
}

func StartServerFast(accessKey string, processor func(function string, parameter []byte) (response []byte, err error)) *Server {
	privateKey, _ := utils.GeneratePrivateKey()
	s := StartServer(privateKey, accessKey, processor)
	s.privateKey = privateKey
	return s
}

func (c *Server) Address() common.Address {
	return utils.AddressForPublicKey(&c.privateKey.PublicKey)
}

func (c *Server) Stop() {
	c.serverConnection.Stop()
}

func (c *Server) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == c.accessKey {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *Server) ServerProcessorCall(authData []byte, function string, parameter []byte) (response []byte, err error) {
	if c.processor != nil {
		return c.processor(function, parameter)
	}
	return nil, errors.New("not implemented")
}
