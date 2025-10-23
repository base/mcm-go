package bridge

import (
	"fmt"

	bridgeInstructions "github.com/base/mcm-go/cmd/internal/bridge/instructions"
	bridgeState "github.com/base/mcm-go/cmd/internal/bridge/state"
	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	solana "github.com/gagliardetto/solana-go"
	ucli "github.com/urfave/cli/v2"
)

// PauseCommand returns the pause/unpause bridge proposal creation command
func PauseCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "pause",
		Usage: "Create a proposal to pause or unpause the bridge",
		Flags: append(flags.ProposalCreationFlags(),
			&ucli.StringFlag{
				Name:     "bridge-program-id",
				Usage:    "Bridge program ID (base58)",
				EnvVars:  []string{"BRIDGE_PROGRAM_ID"},
				Required: true,
			},
			&ucli.BoolFlag{
				Name:  "pause",
				Usage: "Pause the bridge",
			},
			&ucli.BoolFlag{
				Name:  "unpause",
				Usage: "Unpause the bridge",
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parsePauseParams(c)
			if err != nil {
				return err
			}

			// Load MCM client first (needed for fetching bridge state)
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Fetch bridge account to get guardian
			fmt.Println("Fetching Bridge account...")
			bridgeFetcher := bridgeState.NewFetcher(mcmClient.RPC, params.bridgeProgramID)
			bridgeAccount, err := bridgeFetcher.GetBridge(c.Context)
			if err != nil {
				return fmt.Errorf("failed to fetch bridge account: %w", err)
			}

			fmt.Printf("Guardian (from bridge): %s\n", bridgeAccount.Guardian)

			// Create set pause status instruction
			pauseIx, err := bridgeInstructions.SetPauseStatus(bridgeInstructions.SetPauseStatusParams{
				Paused:    params.paused,
				Guardian:  bridgeAccount.Guardian,
				ProgramID: params.bridgeProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to create set pause status instruction: %w", err)
			}

			proposalSvc := services.NewProposalService(mcmClient)

			// Create proposal by fetching on-chain state
			fmt.Println("Fetching MCM on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           params.multisigID,
				ValidUntil:           params.validUntil,
				Instructions:         []solana.Instruction{pauseIx},
				OverridePreviousRoot: params.overridePreviousRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			if err := proposalIO.SaveProposal(p, params.outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			action := "pause"
			if !params.paused {
				action = "unpause"
			}
			fmt.Printf("\nBridge %s proposal created successfully and saved to %s\n", action, params.outputPath)
			return nil
		},
	}
}

// pauseParams holds parsed parameters for the pause command
type pauseParams struct {
	// Common MCM parameters
	multisigID           [32]byte
	validUntil           uint32
	overridePreviousRoot bool
	outputPath           string
	// Specific bridge parameters
	bridgeProgramID solana.PublicKey
	paused          bool
}

// parsePauseParams parses and validates CLI parameters for pause command
func parsePauseParams(c *ucli.Context) (*pauseParams, error) {
	// 1. Parse common MCM parameters
	multisigID, err := hex.Parse32(c.String("multisig-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid multisig-id: %w", err)
	}

	validUntil := uint32(c.Uint64("valid-until"))
	overridePreviousRoot := c.Bool("override-previous-root")
	outputPath := c.String("output")

	// 2. Parse specific bridge parameters
	bridgeProgramID, err := solana.PublicKeyFromBase58(c.String("bridge-program-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid bridge program ID: %w", err)
	}

	pause := c.Bool("pause")
	unpause := c.Bool("unpause")

	// Validate: exactly one of --pause or --unpause must be specified
	if pause && unpause {
		return nil, fmt.Errorf("cannot specify both --pause and --unpause")
	}
	if !pause && !unpause {
		return nil, fmt.Errorf("must specify either --pause or --unpause")
	}

	paused := pause // true if --pause, false if --unpause

	// 3. Print summary (common first, then specific)
	fmt.Printf("Multisig ID: 0x%x\n", multisigID)
	fmt.Printf("Bridge program ID: %s\n", bridgeProgramID)
	action := "unpause"
	if paused {
		action = "pause"
	}
	fmt.Printf("Action: %s\n", action)

	// 4. Return with common fields first
	return &pauseParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		bridgeProgramID:      bridgeProgramID,
		paused:               paused,
	}, nil
}
