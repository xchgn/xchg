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

package utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func PublicKeyToBytes(publicKey *ecdsa.PublicKey) (publicKeyDer []byte) {
	if publicKey == nil {
		return
	}
	return crypto.FromECDSAPub(publicKey)
}

func PublicKeyFromBytes(publicKeyDer []byte) (publicKey *ecdsa.PublicKey, err error) {
	pubKey, err := crypto.UnmarshalPubkey(publicKeyDer)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func StringToAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, fmt.Errorf("invalid address: %s", address)
	}
	return common.HexToAddress(address), nil
}
