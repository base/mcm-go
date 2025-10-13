package mcm

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	mcmHex "github.com/base/mcm-go/pkg/hex"
	mcmInstructions "github.com/base/mcm-go/pkg/instructions"
	mcmpda "github.com/base/mcm-go/pkg/pda"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	solana "github.com/gagliardetto/solana-go"
	ucli "github.com/urfave/cli/v2"
)

// AcceptOwnershipCommand returns the accept ownership proposal creation command
func AcceptOwnershipCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "accept-ownership",
		Usage: "Create a proposal to accept ownership of the multisig",
		Flags: flags.ProposalCreationFlags(),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parseAcceptOwnershipParams(c)
			if err != nil {
				return err
			}

			// Load MCM client
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Derive MCM authority
			mcmAuthority, _, err := mcmpda.MultisigSignerPDA(mcmClient.ProgramID, params.multisigID)
			if err != nil {
				return fmt.Errorf("failed to derive multisig authority PDA: %w", err)
			}
			fmt.Printf("MCM authority: %s\n", mcmAuthority)

			// Create accept ownership instruction
			acceptOwnershipIx, err := mcmInstructions.AcceptOwnership(mcmInstructions.AcceptOwnershipParams{
				MultisigID: params.multisigID,
				Authority:  mcmAuthority,
				ProgramID:  mcmClient.ProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to create accept ownership instruction: %w", err)
			}

			// Create proposal using ProposalService
			proposalSvc := services.NewProposalService(mcmClient)

			fmt.Println("Fetching MCM on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           params.multisigID,
				ValidUntil:           params.validUntil,
				Instructions:         []solana.Instruction{acceptOwnershipIx},
				OverridePreviousRoot: params.overridePreviousRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			if err := proposalIO.SaveProposal(p, params.outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			fmt.Printf("\nAccept ownership proposal created successfully and saved to %s\n", params.outputPath)
			fmt.Printf("  Multisig ID: 0x%x\n", p.MultisigID)
			fmt.Printf("  Valid Until: %d\n", p.ValidUntil)
			fmt.Printf("  Instructions: %d\n", len(p.Instructions))
			fmt.Printf("  Chain ID: %d\n", p.RootMetadata.ChainID)
			fmt.Printf("  Pre Op Count: %d\n", p.RootMetadata.PreOpCount)
			fmt.Printf("  Post Op Count: %d\n", p.RootMetadata.PostOpCount)
			return nil
		},
	}
}

// acceptOwnershipParams holds parsed parameters for the accept ownership command
type acceptOwnershipParams struct {
	// Common MCM parameters
	multisigID           [32]byte
	validUntil           uint32
	overridePreviousRoot bool
	outputPath           string
}

// parseAcceptOwnershipParams parses and validates CLI parameters for accept ownership command
func parseAcceptOwnershipParams(c *ucli.Context) (*acceptOwnershipParams, error) {
	// Parse common MCM parameters
	multisigID, err := mcmHex.Parse32(c.String("multisig-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid multisig-id: %w", err)
	}

	validUntil := uint32(c.Uint64("valid-until"))
	overridePreviousRoot := c.Bool("override-previous-root")
	outputPath := c.String("output")

	return &acceptOwnershipParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
	}, nil
}
