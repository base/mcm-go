package util

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/proposal"
	proposalIO "mcm-go/pkg/proposal/io"

	ucli "github.com/urfave/cli/v2"
)

// LoadClient loads the MCM client from CLI flags
func LoadClient(c *ucli.Context) (*client.Client, error) {
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc"),
		WSUrl:       c.String("ws"),
		ProgramID:   c.String("program-id"),
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
