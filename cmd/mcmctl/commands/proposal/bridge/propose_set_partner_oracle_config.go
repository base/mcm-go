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

// SetPartnerOracleConfigCommand returns the set partner oracle config proposal creation command
func SetPartnerOracleConfigCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "set-partner-oracle-config",
		Usage: "Create a proposal to update the partner oracle configuration",
		Flags: append(flags.ProposalCreationFlags(),
			&ucli.StringFlag{
				Name:     "bridge-program-id",
				Usage:    "Bridge program ID (base58)",
				EnvVars:  []string{"BRIDGE_PROGRAM_ID"},
				Required: true,
			},
			&ucli.Uint64Flag{
				Name:     "required-threshold",
				Usage:    "Number of partner signatures required (uint8, 0-255)",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parseSetPartnerOracleConfigParams(c)
			if err != nil {
				return err
			}

			// Load MCM client first (needed for fetching upgrade authority)
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Fetch upgrade authority from program data
			fmt.Println("Fetching upgrade authority...")
			bridgeFetcher := bridgeState.NewFetcher(mcmClient.RPC, params.bridgeProgramID)
			upgradeAuthority, err := bridgeFetcher.GetUpgradeAuthority(c.Context)
			if err != nil {
				return fmt.Errorf("failed to fetch upgrade authority: %w", err)
			}

			fmt.Printf("Upgrade authority (from program data): %s\n", upgradeAuthority)

			// Create set partner oracle config instruction
			setConfigIx, err := bridgeInstructions.SetPartnerOracleConfig(bridgeInstructions.SetPartnerOracleConfigParams{
				RequiredThreshold: params.requiredThreshold,
				UpgradeAuthority:  upgradeAuthority,
				ProgramID:         params.bridgeProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to create set partner oracle config instruction: %w", err)
			}

			proposalSvc := services.NewProposalService(mcmClient)

			// Create proposal by fetching on-chain state
			fmt.Println("Fetching MCM on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           params.multisigID,
				ValidUntil:           params.validUntil,
				Instructions:         []solana.Instruction{setConfigIx},
				OverridePreviousRoot: params.overridePreviousRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			if err := proposalIO.SaveProposal(p, params.outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			fmt.Printf("\nBridge set partner oracle config proposal created successfully and saved to %s\n", params.outputPath)
			return nil
		},
	}
}

// setPartnerOracleConfigParams holds parsed parameters for the set partner oracle config command
type setPartnerOracleConfigParams struct {
	// Common MCM parameters
	multisigID           [32]byte
	validUntil           uint32
	overridePreviousRoot bool
	outputPath           string
	// Specific bridge parameters
	bridgeProgramID   solana.PublicKey
	requiredThreshold uint8
}

// parseSetPartnerOracleConfigParams parses and validates CLI parameters for set partner oracle config command
func parseSetPartnerOracleConfigParams(c *ucli.Context) (*setPartnerOracleConfigParams, error) {
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

	requiredThresholdVal := c.Uint64("required-threshold")
	if requiredThresholdVal > 255 {
		return nil, fmt.Errorf("required-threshold must be between 0 and 255, got %d", requiredThresholdVal)
	}
	requiredThreshold := uint8(requiredThresholdVal)

	// 3. Print summary (common first, then specific)
	fmt.Printf("Multisig ID: 0x%x\n", multisigID)
	fmt.Printf("Bridge program ID: %s\n", bridgeProgramID)
	fmt.Printf("Required threshold: %d\n", requiredThreshold)

	// 4. Return with common fields first
	return &setPartnerOracleConfigParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		bridgeProgramID:      bridgeProgramID,
		requiredThreshold:    requiredThreshold,
	}, nil
}
