package utils_test

import (
	"bytes"
	"testing"

	"github.com/xchgn/xchg/utils"
)

func TestPack(t *testing.T) {
	testTable := []struct {
		data []byte
	}{
		{
			data: []byte{},
		},
		{
			data: []byte{0},
		},
		{
			data: []byte{0, 1, 2, 3},
		},
		{
			data: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, testCase := range testTable {
		t.Log("Data:", testCase.data)
		packedData := utils.Pack(testCase.data)
		t.Log("Packed:", packedData)
		unpackedData, err := utils.Unpack(packedData)
		if err != nil {
			t.Error("error", err)
		}
		t.Log("Unpacked:", unpackedData)
		if !bytes.Equal(testCase.data, unpackedData) {
			t.Error("error")
		}
	}
}
