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

// WithMessageHash computes the EIP-712 message hash that signers need to sign and returns a ProposalWithMessageHash
func (pwr *ProposalWithRoot) WithMessageHash(programID solana.PublicKey) (*ProposalWithMessageHash, error) {
	if pwr.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}

	// Convert programID to [32]byte
	var programID32 [32]byte
	copy(programID32[:], programID[:])

	// Build the EIP-712 typed data structure
	typedData := buildEIP712TypedData(
		pwr.Root,
		pwr.ValidUntil,
		pwr.RootMetadata.ChainID,
		programID32,
	)

	// Hash the domain
	domainSepBytes, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain: %w", err)
	}
	var domainSeparator [32]byte
	copy(domainSeparator[:], domainSepBytes)

	// Hash the message
	structHashBytes, err := typedData.HashStruct("RootValidation", typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash struct: %w", err)
	}
	var structHash [32]byte
	copy(structHash[:], structHashBytes)

	// Compute final EIP-712 hash: \x19\x01 || domainSeparator || structHash
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator[:]...)
	rawData = append(rawData, structHash[:]...)

	hash := crypto.Keccak256Hash(rawData)
	messageHash := [32]byte(hash)

	return &ProposalWithMessageHash{
		ProposalWithRoot: pwr,
		MessageHash:      messageHash,
		DomainSeparator:  domainSeparator,
		StructHash:       structHash,
		TypedData:        typedData,
	}, nil
}

// buildEIP712TypedData constructs the complete EIP-712 typed data structure
func buildEIP712TypedData(
	root [32]byte,
	validUntil uint32,
	chainID uint64,
	programID [32]byte,
) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "salt", Type: "bytes32"},
			},
			"RootValidation": []apitypes.Type{
				{Name: "root", Type: "bytes32"},
				{Name: "validUntil", Type: "uint32"},
			},
		},
		PrimaryType: "RootValidation",
		Domain: apitypes.TypedDataDomain{
			Name:    "ManyChainMultiSig",
			Version: "1",
			ChainId: (*math.HexOrDecimal256)(big.NewInt(int64(chainID))),
			Salt:    hexutil.Encode(programID[:]),
		},
		Message: apitypes.TypedDataMessage{
			"root":       hexutil.Encode(root[:]),
			"validUntil": fmt.Sprintf("%d", validUntil),
		},
	}
}
