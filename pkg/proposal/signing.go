package proposal

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// WithHashToSign computes the hash that signers need to sign and returns a ProposalToSign
func (pwr *ProposalWithRoot) WithHashToSign() (*ProposalToSign, error) {
	if pwr.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}

	hashToSign := computeHashToSign(pwr.Root, pwr.ValidUntil)

	return &ProposalToSign{
		ProposalWithRoot: pwr,
		HashToSign:       hashToSign,
	}, nil
}

// computeHashToSign computes the hash that signers need to sign
func computeHashToSign(root [32]byte, validUntil uint32) [32]byte {
	// Hash to sign: keccak256(root || validUntil)
	// validUntil is left-padded to 32 bytes in big-endian for EVM compatibility
	data := make([]byte, 64) // 32 bytes root + 32 bytes validUntil
	copy(data[:32], root[:])
	binary.BigEndian.PutUint32(data[60:], validUntil) // offset 60 = 32 + 28

	return crypto.Keccak256Hash(data)
}
