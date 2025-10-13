package bridge

import (
	"bytes"
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	bin "github.com/gagliardetto/binary"
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
			&ucli.StringFlag{
				Name:     "bridge",
				Usage:    "Bridge account address",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "guardian",
				Usage:    "Guardian account address (MCM authority)",
				Required: true,
			},
			&ucli.BoolFlag{
				Name:     "paused",
				Usage:    "Set to true to pause the bridge, false to unpause",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parsePauseParams(c)
			if err != nil {
				return err
			}

			// Create set pause status instruction
			pauseIx, err := newSetPauseStatusInstruction(
				params.paused,
				params.bridge,
				params.guardian,
				params.bridgeProgramID,
			)
			if err != nil {
				return fmt.Errorf("failed to create set pause status instruction: %w", err)
			}

			// Load MCM client and create proposal service
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

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
	bridge          solana.PublicKey
	guardian        solana.PublicKey
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

	bridge, err := solana.PublicKeyFromBase58(c.String("bridge"))
	if err != nil {
		return nil, fmt.Errorf("invalid bridge address: %w", err)
	}

	guardian, err := solana.PublicKeyFromBase58(c.String("guardian"))
	if err != nil {
		return nil, fmt.Errorf("invalid guardian address: %w", err)
	}

	paused := c.Bool("paused")

	// 3. Print summary (common first, then specific)
	fmt.Printf("Multisig ID: 0x%x\n", multisigID)
	fmt.Printf("Bridge program ID: %s\n", bridgeProgramID)
	fmt.Printf("Bridge account: %s\n", bridge)
	fmt.Printf("Guardian account: %s\n", guardian)
	fmt.Printf("Pause status: %v\n", paused)

	// 4. Return with common fields first
	return &pauseParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		bridgeProgramID:      bridgeProgramID,
		bridge:               bridge,
		guardian:             guardian,
		paused:               paused,
	}, nil
}

// Instruction discriminator for set_pause_status
var instructionSetPauseStatus = [8]byte{118, 25, 145, 217, 114, 209, 236, 145}

// newSetPauseStatusInstruction builds a "set_pause_status" instruction for the bridge program.
// Set the pause status for the bridge. Only the guardian can call this function.
// This function is mostly based on the generated code from the bridge program IDL with anchor-go.
//
// # Arguments
// * `newPaused` - The new pause status (true for paused, false for unpaused)
// * `bridgeAccount` - The bridge account containing configuration
// * `guardianAccount` - The guardian account authorized to update configuration
// * `bridgeProgramID` - The bridge program ID
func newSetPauseStatusInstruction(
	newPaused bool,
	bridgeAccount solana.PublicKey,
	guardianAccount solana.PublicKey,
	bridgeProgramID solana.PublicKey,
) (solana.Instruction, error) {
	buf := new(bytes.Buffer)
	enc := bin.NewBorshEncoder(buf)

	// Encode the instruction discriminator
	err := enc.WriteBytes(instructionSetPauseStatus[:], false)
	if err != nil {
		return nil, fmt.Errorf("failed to write instruction discriminator: %w", err)
	}

	// Serialize the newPaused parameter
	err = enc.Encode(newPaused)
	if err != nil {
		return nil, fmt.Errorf("failed to encode newPaused parameter: %w", err)
	}

	accounts := solana.AccountMetaSlice{}

	// Account 0: Bridge account (writable)
	accounts.Append(solana.NewAccountMeta(bridgeAccount, true, false))
	// Account 1: Guardian account (signer)
	accounts.Append(solana.NewAccountMeta(guardianAccount, false, true))

	// Create the instruction
	return solana.NewInstruction(
		bridgeProgramID,
		accounts,
		buf.Bytes(),
	), nil
}
