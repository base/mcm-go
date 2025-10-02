package crypto

import (
	"fmt"
)

// ECDSASignature represents an ECDSA signature with components used in Ethereum signature verification
type ECDSASignature struct {
	V uint8
	R [32]byte
	S [32]byte
}

// DecodeSignature decodes a 65-byte ECDSA signature into its components
func DecodeSignature(sig []byte) (*ECDSASignature, error) {
	if len(sig) != 65 {
		return nil, fmt.Errorf("signature must be 65 bytes, got %d", len(sig))
	}

	var r, s [32]byte
	copy(r[:], sig[0:32])
	copy(s[:], sig[32:64])
	v := sig[64]

	return &ECDSASignature{
		V: v,
		R: r,
		S: s,
	}, nil
}

// Encode encodes the signature back to a 65-byte slice
func (sig *ECDSASignature) Encode() []byte {
	result := make([]byte, 65)
	copy(result[0:32], sig.R[:])
	copy(result[32:64], sig.S[:])
	result[64] = sig.V
	return result
}
