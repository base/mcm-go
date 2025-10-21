package proposal

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gagliardetto/solana-go"
)

// WithHashToSign computes the hash that signers need to sign using EIP-712 and returns a ProposalToSign
func (pwr *ProposalWithRoot) WithHashToSign(programID solana.PublicKey) (*ProposalToSign, error) {
	if pwr.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}

	// Convert programID to [32]byte
	var programID32 [32]byte
	copy(programID32[:], programID[:])

	// Compute EIP-712 hash
	hashToSign, err := computeEIP712HashToSign(
		pwr.Root,
		pwr.ValidUntil,
		pwr.RootMetadata.ChainID,
		programID32,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to compute EIP-712 hash: %w", err)
	}

	return &ProposalToSign{
		ProposalWithRoot: pwr,
		HashToSign:       hashToSign,
	}, nil
}

// computeEIP712HashToSign computes the EIP-712 hash that signers need to sign
// Following the EIP-712 specification with:
// - Domain: ManyChainMultiSig v1 (with chainId, verifyingContract=0x00...00, salt=programId)
// - Message: RootValidation(bytes32 root, uint32 validUntil)
// - Hash: keccak256("\x19\x01" || domainSeparator || structHash)
func computeEIP712HashToSign(
	root [32]byte,
	validUntil uint32,
	chainID uint64,
	programID [32]byte,
) ([32]byte, error) {
	// Define the typed data structure
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
				{Name: "salt", Type: "bytes32"},
			},
			"RootValidation": []apitypes.Type{
				{Name: "root", Type: "bytes32"},
				{Name: "validUntil", Type: "uint32"},
			},
		},
		PrimaryType: "RootValidation",
		Domain: apitypes.TypedDataDomain{
			Name:              "ManyChainMultiSig",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(big.NewInt(int64(chainID))),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
			Salt:              hexutil.Encode(programID[:]),
		},
		Message: apitypes.TypedDataMessage{
			"root":       hexutil.Encode(root[:]),
			"validUntil": fmt.Sprintf("%d", validUntil),
		},
	}

	// Hash the domain
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to hash domain: %w", err)
	}

	// Hash the message
	structHash, err := typedData.HashStruct("RootValidation", typedData.Message)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to hash struct: %w", err)
	}

	// Compute final EIP-712 hash: \x19\x01 || domainSeparator || structHash
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, structHash...)

	hash := crypto.Keccak256Hash(rawData)
	return [32]byte(hash), nil
}
