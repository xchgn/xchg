package utils

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

const Base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

const AddressBytesSize = 20

// EncryptBytesWithPublicKey encrypts data with the given ECIES public key.
func EncryptBytesWithPublicKey(pubKey *ecdsa.PublicKey, data []byte) ([]byte, error) {
	eciesPubKey := ecies.ImportECDSAPublic(pubKey)
	encryptedData, err := ecies.Encrypt(rand.Reader, eciesPubKey, data, nil, nil)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}

// DecryptBytesWithPrivateKey decrypts data with the given ECIES private key.
func DecryptBytesWithPrivateKey(privKey *ecdsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	eciesPrivKey := ecies.ImportECDSA(privKey)
	decryptedData, err := eciesPrivKey.Decrypt(encryptedData, nil, nil)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

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
func VerifySignature(pubKey *ecdsa.PublicKey, data []byte, signature []byte) bool {
	hash := crypto.Keccak256Hash(data)
	return crypto.VerifySignature(crypto.FromECDSAPub(pubKey), hash.Bytes(), signature[:len(signature)-1]) // -1 to remove the recovery id
}

func AddressBSForPublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.New("publicKey == nil")
	}
	bs := PublicKeyToDer(publicKey)
	return AddressBSForPublicKeyBS(bs)
}

func AddressForPublicKey(publicKey *ecdsa.PublicKey) common.Address {
	if publicKey == nil {
		return common.Address{}
	}
	return crypto.PubkeyToAddress(*publicKey)
}

func AddressForPublicKeyBS(publicKeyBS []byte) string {
	if len(publicKeyBS) == 0 {
		return ""
	}

	hash := publicKeyBS
	for i := 0; i < 1; i++ {
		h := sha256.Sum256(hash)
		hash = h[:]
	}

	return "#" + strings.ToLower(base32.StdEncoding.EncodeToString(hash[:AddressBytesSize]))
}

func BytesToAddress(bytes []byte) (common.Address, error) {
	if len(bytes) != 20 {
		return common.Address{}, fmt.Errorf("invalid length: expected 20 bytes, got %d bytes", len(bytes))
	}
	address := common.BytesToAddress(bytes)
	return address, nil
}

func AddressBSForPublicKeyBS(publicKeyBS []byte) ([]byte, error) {
	pubKey, err := crypto.UnmarshalPubkey(publicKeyBS)
	if err != nil {
		return common.Address{}.Bytes(), err
	}
	return crypto.PubkeyToAddress(*pubKey).Bytes(), nil
}

func PublicKeyToDer(publicKey *ecdsa.PublicKey) (publicKeyDer []byte) {
	if publicKey == nil {
		return
	}
	return crypto.FromECDSAPub(publicKey)
}

func PublicKeyFromDer(publicKeyDer []byte) (publicKey *ecdsa.PublicKey, err error) {
	pubKey, err := crypto.UnmarshalPubkey(publicKeyDer)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func DecryptAESGCM(encryptedMessage []byte, key []byte) (decryptedMessage []byte, err error) {
	ch, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(ch)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedMessage) < nonceSize {
		return nil, errors.New("wrong nonce")
	}
	nonce, ciphertext := encryptedMessage[:nonceSize], encryptedMessage[nonceSize:]
	decryptedMessage, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return
}

func EncryptAESGCM(decryptedMessage []byte, key []byte) (encryptedMessage []byte, err error) {
	var ch cipher.Block
	ch, err = aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(ch)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	encryptedMessage = gcm.Seal(nonce, nonce, decryptedMessage, nil)
	return
}

func PackBytes(data []byte) []byte {
	var err error
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	var zipFile io.Writer
	zipFile, err = zipWriter.Create("data")
	if err == nil {
		_, err = zipFile.Write(data)
	}
	err = zipWriter.Close()
	return buf.Bytes()
}

func UnpackBytes(zippedData []byte) (result []byte, err error) {
	buf := bytes.NewReader(zippedData)
	var zipFile *zip.Reader
	zipFile, err = zip.NewReader(buf, buf.Size())
	if err != nil {
		return
	}
	var file fs.File
	file, err = zipFile.Open("data")
	if err == nil {
		result, err = ioutil.ReadAll(file)
		_ = file.Close()
	}
	return
}