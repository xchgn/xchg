package aes_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/xchgn/xchg/utils"
)

func TestAES(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	data := []byte{0, 0, 0, 0, 0}
	t.Log("data:", data)

	encryptedData, err := utils.EncryptAESGCM(data, key)
	t.Log("encryptedData:", encryptedData)
	if err != nil {
		t.Error("Encrypt error:", err)
	}

	decryptedData, err := utils.DecryptAESGCM(encryptedData, key)
	t.Log("decryptedData:", decryptedData)
	if err != nil {
		t.Error("Decrypt error:", err)
	}

	if !bytes.Equal(data, decryptedData) {
		t.Error("---")
	}
}
