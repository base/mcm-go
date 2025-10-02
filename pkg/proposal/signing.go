package proposal

import (
	"encoding/binary"
	"fmt"

	"mcm-go/pkg/crypto"
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
	data := make([]byte, 0, 32+4)
	data = append(data, root[:]...)

	validUntilBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(validUntilBytes, validUntil)
	data = append(data, validUntilBytes...)

	return crypto.Keccak256Hash(data)
}
