package blockchain

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/xchgn/suigo/client"
	"github.com/xchgn/xchg/logger"
	"github.com/xchgn/xchg/utils"
)

type RouterAccount struct {
	bc        *Blockchain
	suiClient *client.Client

	xchgPrivateKey ed25519.PrivateKey
	xchgPublicKey  ed25519.PublicKey
	xchgAddress    string
}

func NewRouterAccount() *RouterAccount {
	var c RouterAccount

	c.bc = NewBlockchain()

	// Initialize Sui client
	c.suiClient = client.NewClient(client.TESTNET_URL)
	c.suiClient.InitAccountFromFile("private/sui_seed_phrase.txt")

	// Initialize Xchg keys
	exePath := logger.CurrentExePath()
	xchgSeedPhrase, _ := os.ReadFile(exePath + "/private/xchg_seed_phrase.txt")
	c.xchgPrivateKey, _ = utils.PrivateKeyFromMnemonic(string(xchgSeedPhrase))
	c.xchgPublicKey = c.xchgPrivateKey.Public().(ed25519.PublicKey)
	c.xchgAddress = "0x" + hex.EncodeToString(c.xchgPublicKey)

	return &c
}

func (c *RouterAccount) GetXchgAddress() string {
	return c.xchgAddress
}

func (c *RouterAccount) SuiClient() *client.Client {
	return c.suiClient
}

func (c *RouterAccount) GenerateCheques() {
	// public fun get_cheques_ids(f: &mut Fund, xchgAddressOfRouter: address, count: u32, clock: &Clock, _ctx: &mut TxContext) : vector<address> {

	var p client.MoveCallParameters
	p.PackageId = PACKAGE_ID
	p.ModuleName = "fund"
	p.FunctionName = "get_cheques_ids"
	p.Arguments = []interface{}{
		client.ArgSharedObject(FUND_OBJECT_ID),
		client.ArgAddress(c.GetXchgAddress()),
		client.ArgU32(10),
		client.ArgSharedObject(client.CLOCK_OBJECT_ID),
	}
	res, err := c.suiClient.ExecMoveCall(p, 1000)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("RESULT:", res)
}

type Item struct {
	ChequeId []byte
	AppId    []byte
}

func (c *RouterAccount) ApplyChecques(items []Item) {
	// public fun apply_cheque(f: &mut Fund, pk: vector<u8>,  msg: vector<u8>, sig: vector<u8>, clock: &Clock, _ctx: &mut TxContext) {

	tb := client.NewTransactionBuilder(c.suiClient)
	for _, item := range items {
		pk := make([]byte, 32)
		copy(pk, c.xchgPublicKey)
		msg := make([]byte, 104)
		copy(msg, item.ChequeId)
		copy(msg[32:], c.xchgPublicKey)
		copy(msg[64:64+32], item.AppId)
		binary.LittleEndian.PutUint64(msg[96:], 1)

		sig := ed25519.Sign(c.xchgPrivateKey, msg)

		cmd := client.NewTransactionBuilderMoveCall()
		cmd.PackageId = PACKAGE_ID
		cmd.ModuleName = "fund"
		cmd.FunctionName = "apply_cheque"
		cmd.Arguments = []interface{}{
			client.ArgSharedObject(FUND_OBJECT_ID),
			client.ArgVecU8(pk),
			client.ArgVecU8(msg),
			client.ArgVecU8(sig),
			client.ArgSharedObject(client.CLOCK_OBJECT_ID),
		}
		tb.AddCommand(cmd)
	}
	res, err := c.suiClient.ExecPTB(tb, client.DEFAULT_GAS_PRICE)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("RESULT:", res)
}

func (c *RouterAccount) FetchRouterObject() (*RouterObject, error) {
	return c.bc.GetRouterObject(c.GetXchgAddress())
}
