package proposal

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// CreateCommand returns the proposal create command
func CreateCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "create",
		Usage: "Create a proposal from instructions and on-chain state, then save it to a file",
		Flags: append(flags.OnchainReadFlags(),
			&ucli.StringFlag{
				Name:     "instructions",
				Aliases:  []string{"i"},
				Usage:    "Path to instructions JSON file (simplified format with only instructions array)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.Uint64Flag{
				Name:     "valid-until",
				Usage:    "Proposal expiration timestamp (Unix timestamp)",
				Required: true,
			},
			&ucli.BoolFlag{
				Name:  "override-previous-root",
				Usage: "Override previous Merkle root (invalidates pending operations)",
				Value: false,
			},
			&ucli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output file path for the generated proposal",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			// Load instructions from simplified JSON file
			instructionsPath := c.String("instructions")
			instructions, err := proposalIO.LoadInstructions(instructionsPath)
			if err != nil {
				return fmt.Errorf("failed to load instructions: %w", err)
			}

			fmt.Printf("Loaded %d instruction(s) from %s\n", len(instructions), instructionsPath)

			// Parse multisig ID
			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Load client and create proposal service
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			proposalSvc := services.NewProposalService(mcmClient)

			// Create proposal by fetching on-chain state
			fmt.Println("Fetching on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           multisigID,
				ValidUntil:           uint32(c.Uint64("valid-until")),
				Instructions:         instructions,
				OverridePreviousRoot: c.Bool("override-previous-root"),
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			outputPath := c.String("output")
			if err := proposalIO.SaveProposal(p, outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			fmt.Printf("\nProposal created successfully and saved to %s\n", outputPath)
			fmt.Printf("  Multisig ID: 0x%x\n", p.MultisigID)
			fmt.Printf("  Valid Until: %d\n", p.ValidUntil)
			fmt.Printf("  Instructions: %d\n", len(p.Instructions))
			fmt.Printf("  Chain ID: %d\n", p.RootMetadata.ChainID)
			fmt.Printf("  Multisig: %s\n", p.RootMetadata.Multisig)
			fmt.Printf("  Pre Op Count: %d\n", p.RootMetadata.PreOpCount)
			fmt.Printf("  Post Op Count: %d\n", p.RootMetadata.PostOpCount)
			fmt.Printf("  Override Previous Root: %v\n", p.RootMetadata.OverridePreviousRoot)

			return nil
		},
	}
}
