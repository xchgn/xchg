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

	"github.com/ethereum/go-ethereum/crypto"
)

// SignData signs the given data using the provided private key.
func SignData(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// VerifySignature verifies the signature of the given data using the provided public key.
/*func VerifySignature(pubKey *ecdsa.PublicKey, data []byte, signature []byte) bool {
	hash := crypto.Keccak256Hash(data)
	return crypto.VerifySignature(crypto.FromECDSAPub(pubKey), hash.Bytes(), signature[:len(signature)-1]) // -1 to remove the recovery id
}

func RecoverPublicKeyFromSignature(message, signature []byte) (*ecdsa.PublicKey, error) {
	publicKeyAsBytes, err := crypto.Ecrecover(message, signature)
	if err != nil {
		return nil, err
	}
	recoveredPubKey, err := crypto.UnmarshalPubkey(publicKeyAsBytes)
	return recoveredPubKey, err
}
*/
