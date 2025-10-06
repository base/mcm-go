package crypto

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// RecoverAddressFromSig recovers the address from a signature. Accepts 27/28 recovery id and normalizes to 0/1.
func RecoverAddressFromSig(hash [32]byte, sig [65]byte) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: got %d, want 65", len(sig))
	}

	tmp := make([]byte, 65)
	copy(tmp, sig[:])
	if tmp[64] >= 27 {
		tmp[64] -= 27
	}

	pub, err := crypto.SigToPub(hash[:], tmp)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

// MessageHash computes the hash of a message in the format of "\x19Ethereum Signed Message:\n32" + keccak256(message)
func MessageHash(msg []byte) [32]byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msg))
	data := append([]byte(prefix), msg...)
	return crypto.Keccak256Hash(data)
}
