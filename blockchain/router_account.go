package blockchain

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"

	"github.com/xchgn/suigo/client"
	"github.com/xchgn/xchg/logger"
	"github.com/xchgn/xchg/utils"
)

type RouterAccount struct {
	suiClient *client.Client

	xchgPrivateKey ed25519.PrivateKey
	xchgPublicKey  ed25519.PublicKey
	xchgAddress    string
}

func NewRouterAccount() *RouterAccount {
	var c RouterAccount

	// Initialize Sui client
	c.suiClient = client.NewClient(client.TESTNET_URL)
	c.suiClient.InitAccountFromFile("private/sui_seed_phrase.txt")

	// Initialize Xchg keys
	exePath := logger.CurrentExePath()
	xchgSeedPhrase, _ := os.ReadFile(exePath + "/private/xchg_seed_phrase.txt")
	c.xchgPrivateKey, _ = utils.PrivateKeyFromMnemonic(string(xchgSeedPhrase))
	c.xchgPublicKey = c.xchgPrivateKey.Public().(ed25519.PublicKey)
	c.xchgAddress = hex.EncodeToString(c.xchgPublicKey)

	return &c
}

func (c *RouterAccount) GetXchgAddress() string {
	return c.xchgAddress
}

func (c *RouterAccount) SuiClient() *client.Client {
	return c.suiClient
}
