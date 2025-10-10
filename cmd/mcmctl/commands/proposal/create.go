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
			flags.MultisigIDFlag(),
			flags.ValidUntilFlag(),
			flags.InstructionsFlag(),
			flags.OverridePreviousRootFlag(),
			flags.OutputFlag(),
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

			fmt.Printf("Proposal created successfully and saved to %s\n", outputPath)
			return nil
		},
	}
}
