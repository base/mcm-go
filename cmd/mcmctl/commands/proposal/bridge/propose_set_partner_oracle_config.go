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

// CreateSetPartnerOracleConfigCommand returns the set partner oracle config proposal creation command
func CreateSetPartnerOracleConfigCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "create-set-partner-oracle-config",
		Usage: "Create a proposal to update the partner oracle configuration",
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

			// Create set partner oracle config instruction
			setConfigIx, err := newSetPartnerOracleConfigInstruction(
				params.requiredThreshold,
				params.bridge,
				params.guardian,
				params.bridgeProgramID,
			)
			if err != nil {
				return fmt.Errorf("failed to create set partner oracle config instruction: %w", err)
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
	bridge            solana.PublicKey
	guardian          solana.PublicKey
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

	bridge, err := solana.PublicKeyFromBase58(c.String("bridge"))
	if err != nil {
		return nil, fmt.Errorf("invalid bridge address: %w", err)
	}

	guardian, err := solana.PublicKeyFromBase58(c.String("guardian"))
	if err != nil {
		return nil, fmt.Errorf("invalid guardian address: %w", err)
	}

	requiredThresholdVal := c.Uint64("required-threshold")
	if requiredThresholdVal > 255 {
		return nil, fmt.Errorf("required-threshold must be between 0 and 255, got %d", requiredThresholdVal)
	}
	requiredThreshold := uint8(requiredThresholdVal)

	// 3. Print summary (common first, then specific)
	fmt.Printf("Multisig ID: 0x%x\n", multisigID)
	fmt.Printf("Bridge program ID: %s\n", bridgeProgramID)
	fmt.Printf("Bridge account: %s\n", bridge)
	fmt.Printf("Guardian account: %s\n", guardian)
	fmt.Printf("Required threshold: %d\n", requiredThreshold)

	// 4. Return with common fields first
	return &setPartnerOracleConfigParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		bridgeProgramID:      bridgeProgramID,
		bridge:               bridge,
		guardian:             guardian,
		requiredThreshold:    requiredThreshold,
	}, nil
}

// Instruction discriminator for set_partner_oracle_config
var instructionSetPartnerOracleConfig = [8]byte{34, 48, 231, 135, 42, 113, 217, 157}

// PartnerOracleConfig represents the partner oracle configuration
type PartnerOracleConfig struct {
	RequiredThreshold uint8
}

// newSetPartnerOracleConfigInstruction builds a "set_partner_oracle_config" instruction for the bridge program.
// Update the partner oracle configuration containing the required signature threshold.
// This function is mostly based on the generated code from the bridge program IDL with anchor-go.
//
// # Arguments
// * `requiredThreshold` - Number of required partner signatures
// * `bridgeAccount` - The bridge account containing configuration
// * `guardianAccount` - The guardian account authorized to update configuration
// * `bridgeProgramID` - The bridge program ID
func newSetPartnerOracleConfigInstruction(
	requiredThreshold uint8,
	bridgeAccount solana.PublicKey,
	guardianAccount solana.PublicKey,
	bridgeProgramID solana.PublicKey,
) (solana.Instruction, error) {
	buf := new(bytes.Buffer)
	enc := bin.NewBorshEncoder(buf)

	// Encode the instruction discriminator
	err := enc.WriteBytes(instructionSetPartnerOracleConfig[:], false)
	if err != nil {
		return nil, fmt.Errorf("failed to write instruction discriminator: %w", err)
	}

	// Serialize the PartnerOracleConfig parameter
	config := PartnerOracleConfig{
		RequiredThreshold: requiredThreshold,
	}
	err = enc.Encode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to encode partner oracle config parameter: %w", err)
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
