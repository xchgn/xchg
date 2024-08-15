package utils

import (
	"encoding/hex"
	"fmt"
)

func TransactionSummary(frame []byte) string {
	if len(frame) < 128 {
		return "wrong frame length"
	}

	tp := ""
	{
		switch frame[4] {
		case 0x10:
			tp = "CR"
		case 0x11:
			tp = "cr"
		case 0x20:
			tp = "PR"
		case 0x21:
			tp = "pr"
		}
	}

	addressDest := frame[64 : 64+32]
	addressSrc := frame[32 : 32+32]
	comment := frame[96:128]

	for i := 0; i < 32; i++ {
		if comment[i] == 0 {
			comment[i] = ' '
		}
	}

	res := fmt.Sprint("ROUTER PUT ", tp, " func ", string(comment), " from ", hex.EncodeToString(addressSrc[:4]), " to ", hex.EncodeToString(addressDest[:4]))
	return res
}
