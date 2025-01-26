package utils

import (
	"crypto/ed25519"

	"github.com/xchgn/suigo/client"
	"github.com/xchgn/suigo/utils/bip39"
)

func PrivateKeyFromMnemonic(mnemonic string) (ed25519.PrivateKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	key, err := client.DeriveForPath("m/44'/784'/0'/0'/0'", seed)
	if err != nil {
		return nil, err
	}

	priKey := ed25519.NewKeyFromSeed(key.Key)
	return priKey, nil
}
