package proposal

import (
	"github.com/gagliardetto/solana-go"
)

// Builder helps construct proposals
type Builder struct {
	proposal *Proposal
}

// NewBuilder creates a new proposal builder
func NewBuilder(multisigID [32]byte, validUntil uint32) *Builder {
	return &Builder{
		proposal: &Proposal{
			MultisigID:   multisigID,
			ValidUntil:   validUntil,
			Instructions: make([]*solana.GenericInstruction, 0),
		},
	}
}

// SetRootMetadata sets the root metadata for the proposal
func (b *Builder) SetRootMetadata(metadata RootMetadata) *Builder {
	b.proposal.RootMetadata = metadata
	return b
}

// AddInstruction adds an instruction to the proposal
func (b *Builder) AddInstruction(ix *solana.GenericInstruction) *Builder {
	b.proposal.Instructions = append(b.proposal.Instructions, ix)
	return b
}

// Build constructs and returns the proposal
func (b *Builder) Build() (*Proposal, error) {
	if err := b.proposal.Validate(); err != nil {
		return nil, err
	}
	return b.proposal, nil
}
