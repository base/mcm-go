package proposal

import (
	"fmt"

	"mcm-go/cmd/mcmctl/flags"
	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	proposalIO "mcm-go/pkg/proposal/io"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// SetRootCommand returns the proposal set-root command
func SetRootCommand() *ucli.Command {
	txFlags := flags.TransactionFlags()

	return &ucli.Command{
		Name:  "set-root",
		Usage: "Load a proposal and set its root on-chain",
		Flags: append([]ucli.Flag{
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
		}, txFlags...),
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")

			// Load proposal from file
			p, err := proposalIO.LoadProposal(filePath)
			if err != nil {
				return fmt.Errorf("failed to load proposal: %w", err)
			}

			// Compute root and proofs
			pwr, err := p.WithRoot()
			if err != nil {
				return fmt.Errorf("failed to compute Merkle root: %w", err)
			}

			// Load client
			mcmClient, err := loadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Set root on-chain
			proposalSvc := services.NewProposalService(mcmClient)
			sig, err := proposalSvc.SetRoot(c.Context, services.SetRootParams{
				MultisigID: p.MultisigID,
				Proposal:   pwr,
			})
			if err != nil {
				return fmt.Errorf("failed to set root: %w", err)
			}

			fmt.Printf("Root set successfully\n")
			fmt.Printf("  Multisig ID: 0x%x\n", p.MultisigID)
			fmt.Printf("  Root: 0x%x\n", pwr.Root)
			fmt.Printf("  Valid Until: %d\n", pwr.ValidUntil)
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

// loadClient loads the MCM client from CLI flags
func loadClient(c *ucli.Context) (*client.Client, error) {
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
