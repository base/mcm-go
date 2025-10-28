package mcm

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/client"
	mcmHex "github.com/base/mcm-go/pkg/hex"
	mcmInstructions "github.com/base/mcm-go/pkg/instructions"
	mcmpda "github.com/base/mcm-go/pkg/pda"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	solana "github.com/gagliardetto/solana-go"
	ucli "github.com/urfave/cli/v2"
)

// UpdateSignersCommand returns the update signers proposal creation command
func UpdateSignersCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "update-signers",
		Usage: "Create a proposal to update MCM signers configuration (init, append, finalize, setConfig)",
		Flags: append(flags.ProposalCreationFlags(),
			&ucli.StringFlag{
				Name:     "new-signers",
				Usage:    "Comma-separated list of new signer addresses (20 bytes hex string, with 0x prefix)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signer-groups",
				Usage:    "Comma-separated list of group indices for each signer (e.g., '0,0,1,1')",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "group-quorums",
				Usage:    "Comma-separated list of quorum thresholds for each group (e.g., '2,3')",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "group-parents",
				Usage:    "Comma-separated list of parent group indices (e.g., '0,0')",
				Required: true,
			},
			&ucli.BoolFlag{
				Name:  "clear-root",
				Usage: "Clear existing root when setting config",
				Value: false,
			},
			&ucli.BoolFlag{
				Name:  "clear-signers",
				Usage: "Clear existing signers before updating (adds ClearSigners instruction at the beginning)",
				Value: false,
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parseUpdateSignersParams(c)
			if err != nil {
				return err
			}

			// Print parsed configuration
			printUpdateSignersConfig(params)

			// Load MCM client
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Build instructions
			proposalIxs, err := buildUpdateSignersInstructions(mcmClient, params)
			if err != nil {
				return err
			}

			// Create proposal using ProposalService
			proposalSvc := services.NewProposalService(mcmClient)

			fmt.Println("\nFetching MCM on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           params.multisigID,
				ValidUntil:           params.validUntil,
				Instructions:         proposalIxs,
				OverridePreviousRoot: params.overridePreviousRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			if err := proposalIO.SaveProposal(p, params.outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			fmt.Printf("\nUpdate signers proposal created successfully and saved to %s\n", params.outputPath)
			return nil
		},
	}
}

// updateSignersParams holds parsed parameters for the update signers command
type updateSignersParams struct {
	// Common MCM parameters
	multisigID           [32]byte
	validUntil           uint32
	overridePreviousRoot bool
	outputPath           string
	// Specific update-signers parameters
	newSigners   [][20]uint8
	signerGroups []uint8
	groupQuorums [32]uint8
	groupParents [32]uint8
	clearSigners bool
	clearRoot    bool
}

// parseUpdateSignersParams parses and validates CLI parameters for update signers command
func parseUpdateSignersParams(c *ucli.Context) (*updateSignersParams, error) {
	// 1. Parse common MCM parameters
	multisigID, err := mcmHex.Parse32(c.String("multisig-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid multisig-id: %w", err)
	}

	validUntil := uint32(c.Uint64("valid-until"))
	overridePreviousRoot := c.Bool("override-previous-root")
	outputPath := c.String("output")

	// 2. Parse specific update-signers parameters

	// Parse new signers (with automatic sorting)
	newSigners, sorted, err := util.ParseAndSortSigners(c.String("new-signers"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse new-signers: %w", err)
	}
	if sorted {
		fmt.Println("Note: signers were reordered to be strictly increasing")
	}

	// Parse signer groups
	signerGroups, err := util.ParseUint8Slice(c.String("signer-groups"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse signer-groups: %w", err)
	}
	if len(signerGroups) != len(newSigners) {
		return nil, fmt.Errorf("signer-groups length (%d) must match new-signers length (%d)", len(signerGroups), len(newSigners))
	}

	// Parse group quorums
	groupQuorums, err := util.ParseAndPadUint8Array32(c.String("group-quorums"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse group-quorums: %w", err)
	}

	// Parse group parents
	groupParents, err := util.ParseAndPadUint8Array32(c.String("group-parents"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse group-parents: %w", err)
	}

	clearSigners := c.Bool("clear-signers")
	clearRoot := c.Bool("clear-root")

	// 3. Return with common fields first
	return &updateSignersParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		newSigners:           newSigners,
		signerGroups:         signerGroups,
		groupQuorums:         groupQuorums,
		groupParents:         groupParents,
		clearSigners:         clearSigners,
		clearRoot:            clearRoot,
	}, nil
}

// printUpdateSignersConfig prints the parsed configuration
func printUpdateSignersConfig(params *updateSignersParams) {
	fmt.Printf("New signers: %d\n", len(params.newSigners))
	for i, signer := range params.newSigners {
		fmt.Printf("  [%d] 0x%x\n", i, signer)
	}
	fmt.Printf("Signer groups: %v\n", params.signerGroups)
	fmt.Printf("Group quorums: %v\n", params.groupQuorums)
	fmt.Printf("Group parents: %v\n", params.groupParents)
	fmt.Printf("Clear signers: %v\n\n", params.clearSigners)
	fmt.Printf("Clear root: %v\n", params.clearRoot)
}

// buildUpdateSignersInstructions builds all instructions for updating signers
func buildUpdateSignersInstructions(mcmClient *client.Client, params *updateSignersParams) ([]solana.Instruction, error) {
	// Derive MCM authority
	mcmAuthority, _, err := mcmpda.MultisigSignerPDA(mcmClient.ProgramID, params.multisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig authority PDA: %w", err)
	}
	fmt.Printf("MCM authority: %s\n\n", mcmAuthority)

	proposalIxs := make([]solana.Instruction, 0)

	// 0. Clear signers instruction (optional)
	if params.clearSigners {
		fmt.Println("Creating ClearSigners instruction...")
		clearSignersIx, err := mcmInstructions.ClearSigners(mcmInstructions.ClearSignersParams{
			MultisigID: params.multisigID,
			Authority:  mcmAuthority,
			ProgramID:  mcmClient.ProgramID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create clear signers instruction: %w", err)
		}
		proposalIxs = append(proposalIxs, clearSignersIx)
	}

	// 1. Initialize signers instruction
	fmt.Println("Creating InitSigners instruction...")
	initSignersIx, err := mcmInstructions.InitSigners(mcmInstructions.InitSignersParams{
		MultisigID:   params.multisigID,
		TotalSigners: uint8(len(params.newSigners)),
		Authority:    mcmAuthority,
		ProgramID:    mcmClient.ProgramID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create init signers instruction: %w", err)
	}
	proposalIxs = append(proposalIxs, initSignersIx)

	// 2. Append signers instructions
	const maxSignersPerAppend = 10
	fmt.Println("Creating AppendSigners instruction(s)...")
	for i := 0; i < len(params.newSigners); i += maxSignersPerAppend {
		end := min(i+maxSignersPerAppend, len(params.newSigners))
		signersChunk := params.newSigners[i:end]

		appendSignersIx, err := mcmInstructions.AppendSigners(mcmInstructions.AppendSignersParams{
			MultisigID:   params.multisigID,
			SignersBatch: signersChunk,
			Authority:    mcmAuthority,
			ProgramID:    mcmClient.ProgramID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create append signers instruction: %w", err)
		}
		proposalIxs = append(proposalIxs, appendSignersIx)
	}

	// 3. Finalize signers instruction
	fmt.Println("Creating FinalizeSigners instruction...")
	finalizeSignersIx, err := mcmInstructions.FinalizeSigners(mcmInstructions.FinalizeSignersParams{
		MultisigID: params.multisigID,
		Authority:  mcmAuthority,
		ProgramID:  mcmClient.ProgramID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create finalize signers instruction: %w", err)
	}
	proposalIxs = append(proposalIxs, finalizeSignersIx)

	// 4. SetConfig instruction
	fmt.Println("Creating SetConfig instruction...")
	setConfigIx, err := mcmInstructions.SetConfig(mcmInstructions.SetConfigParams{
		MultisigID:   params.multisigID,
		SignerGroups: params.signerGroups,
		GroupQuorums: params.groupQuorums,
		GroupParents: params.groupParents,
		ClearRoot:    params.clearRoot,
		Authority:    mcmAuthority,
		ProgramID:    mcmClient.ProgramID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create set config instruction: %w", err)
	}
	proposalIxs = append(proposalIxs, setConfigIx)

	return proposalIxs, nil
}
