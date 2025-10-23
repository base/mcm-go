package util

import (
	"fmt"

	"github.com/base/mcm-go/pkg/proposal"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"

	"github.com/gagliardetto/solana-go"
)

// LoadProposalWithRoot loads a proposal from file and computes its Merkle root
func LoadProposalWithRoot(filePath string) (*proposal.ProposalWithRoot, error) {
	// Load proposal from file
	p, err := proposalIO.LoadProposal(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load proposal: %w", err)
	}

	// Compute root and proofs
	pwr, err := p.WithRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to compute Merkle root: %w", err)
	}

	return pwr, nil
}

// CreateProposalWithMessageHash creates a ProposalWithMessageHash from a ProposalWithRoot by computing the message hash
func CreateProposalWithMessageHash(pwr *proposal.ProposalWithRoot, programID solana.PublicKey) (*proposal.ProposalWithMessageHash, error) {
	pts, err := pwr.WithMessageHash(programID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute message hash: %w", err)
	}
	return pts, nil
}
