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

// WithMessageHash computes the EIP-712 message hash that signers need to sign and returns a ProposalToSign
func (pwr *ProposalWithRoot) WithMessageHash(programID solana.PublicKey) (*ProposalToSign, error) {
	if pwr.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}

	// Convert programID to [32]byte
	var programID32 [32]byte
	copy(programID32[:], programID[:])

	// Compute EIP-712 hash
	messageHash, domainSep, structHash, err := computeEIP712MessageHash(
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
		MessageHash:      messageHash,
		DomainSeparator:  domainSep,
		StructHash:       structHash,
	}, nil
}

// computeEIP712MessageHash computes the EIP-712 message hash that signers need to sign
// Following the EIP-712 specification with:
// - Domain: ManyChainMultiSig v1 (with chainId, salt=programId)
// - Domain type: EIP712Domain(string name,string version,uint256 chainId,bytes32 salt)
// - Message: RootValidation(bytes32 root, uint32 validUntil)
// - Hash: keccak256("\x19\x01" || domainSeparator || structHash)
//
// Returns:
// - messageHash: Final EIP-712 hash to sign
// - domainSeparator: hashStruct(EIP712Domain)
// - structHash: hashStruct(RootValidation)
func computeEIP712MessageHash(
	root [32]byte,
	validUntil uint32,
	chainID uint64,
	programID [32]byte,
) (messageHash [32]byte, domainSeparator [32]byte, structHash [32]byte, err error) {
	// Define the typed data structure
	typedData := apitypes.TypedData{
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

	// Hash the domain
	domainSepBytes, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return [32]byte{}, [32]byte{}, [32]byte{}, fmt.Errorf("failed to hash domain: %w", err)
	}
	copy(domainSeparator[:], domainSepBytes)

	// Hash the message
	structHashBytes, err := typedData.HashStruct("RootValidation", typedData.Message)
	if err != nil {
		return [32]byte{}, [32]byte{}, [32]byte{}, fmt.Errorf("failed to hash struct: %w", err)
	}
	copy(structHash[:], structHashBytes)

	// Compute final EIP-712 hash: \x19\x01 || domainSeparator || structHash
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator[:]...)
	rawData = append(rawData, structHash[:]...)

	hash := crypto.Keccak256Hash(rawData)
	messageHash = [32]byte(hash)
	return
}
