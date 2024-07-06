package xchg

const (
	XchgMaxFrameSize  = 16 * 1024
	XchgSignatureSize = 65
	XchgPublicKeySize = 65
	XchgAddressSize   = 20
	XchgNonceSize     = 16
	XchgAesKeySize    = 32

	// Frame Type Code
	XchgFrameCallRequest          = 0x10
	XchgFrameCallResponse         = 0x11
	XchgFrameGetPublicKeyRequest  = 0x20
	XchgFrameGetPublicKeyResponse = 0x21
)
