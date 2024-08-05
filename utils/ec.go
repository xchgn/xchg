package utils

import (
	"crypto/ed25519"
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

func GenerateEd25519KeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}

func GenerateCurve25519KeyPair() ([]byte, []byte, error) {
	priv := make([]byte, 32)
	_, err := rand.Read(priv)
	if err != nil {
		return nil, nil, err
	}

	// Применение необходимых модификаций к приватному ключу
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	if err != nil {
		return nil, nil, err
	}

	return priv, pub, nil
}

func ExtractPublicKey(privateKey ed25519.PrivateKey) ed25519.PublicKey {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return publicKey
}

func GetSharedKey(localPrivateKey ed25519.PrivateKey, remotePublicKey ed25519.PublicKey) (result []byte, err error) {
	result, err = curve25519.X25519(localPrivateKey, remotePublicKey)
	return
}

func SignMessage(privateKey ed25519.PrivateKey, message []byte) []byte {
	signature := ed25519.Sign(privateKey, message)
	return signature
}

func VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}
