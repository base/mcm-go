// Package proposal provides types and utilities for creating and managing MCM proposals.
package proposal

import (
	"github.com/base/mcm-go/pkg/crypto"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gagliardetto/solana-go"
)

// RootMetadata contains metadata about a proposal root
type RootMetadata struct {
	ChainID              uint64
	Multisig             solana.PublicKey
	PreOpCount           uint64
	PostOpCount          uint64
	OverridePreviousRoot bool
}

// Proposal represents a complete MCM proposal
type Proposal struct {
	MultisigID   [32]byte
	ValidUntil   uint32
	Instructions []solana.Instruction
	RootMetadata RootMetadata
}

// ProposalRoot contains the Merkle root and proofs for a proposal
type ProposalRoot struct {
	Root            [32]byte
	MetadataProof   crypto.Proof
	OperationProofs []crypto.Proof
}

// ProposalWithRoot combines a Proposal with its Merkle root and proofs
// It embeds *Proposal as a pointer to share the same instance in memory
type ProposalWithRoot struct {
	*Proposal
	ProposalRoot
}

// ProposalWithMessageHash extends ProposalWithRoot with the hash that signers need to sign
type ProposalWithMessageHash struct {
	*ProposalWithRoot
	MessageHash     [32]byte           // Final EIP-712 hash (used for ecrecover)
	DomainSeparator [32]byte           // hashStruct(EIP712Domain)
	StructHash      [32]byte           // hashStruct(RootValidation)
	TypedData       apitypes.TypedData // Complete EIP-712 typed data structure (for external signers)
}

// Validate checks if the proposal is valid
func (p *Proposal) Validate() error {
	if p.ValidUntil == 0 {
		return ErrInvalidValidUntil
	}
	if len(p.Instructions) == 0 {
		return ErrNoInstructions
	}
	if p.RootMetadata.PreOpCount > p.RootMetadata.PostOpCount {
		return ErrInvalidOpCount
	}
	return nil
}
