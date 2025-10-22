package util

import (
	"fmt"

	"github.com/base/mcm-go/pkg/cli"
	"github.com/base/mcm-go/pkg/client"
	"github.com/base/mcm-go/pkg/proposal"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"

	"github.com/gagliardetto/solana-go"
	ucli "github.com/urfave/cli/v2"
)

// LoadClient loads the MCM client from CLI flags
// WS and authority are optional (only needed for write operations)
func LoadClient(c *ucli.Context) (*client.Client, error) {
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc-url"),
		WSUrl:       c.String("ws-url"),
		ProgramID:   c.String("mcm-program-id"),
		KeypairPath: c.String("authority"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return client.New(*cfg)
}

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

// CreateProposalToSign creates a ProposalToSign from a ProposalWithRoot by computing the hash to sign
func CreateProposalToSign(pwr *proposal.ProposalWithRoot, programID solana.PublicKey) (*proposal.ProposalToSign, error) {
	pts, err := pwr.WithMessageHash(programID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash to sign: %w", err)
	}
	return pts, nil
}

// ParseProgramID parses a program ID from a base58 string
func ParseProgramID(programIDStr string) (solana.PublicKey, error) {
	programID, err := solana.PublicKeyFromBase58(programIDStr)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("invalid program ID: %w", err)
	}
	return programID, nil
}
