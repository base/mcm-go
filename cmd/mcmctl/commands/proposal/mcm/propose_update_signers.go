package mcm

import (
	"fmt"
	"strings"

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

// CreateUpdateSignersCommand returns the update signers proposal creation command
func CreateUpdateSignersCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "create-update-signers",
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
	// Parse new signers
	newSigners, err := parseSigners(c.String("new-signers"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse new-signers: %w", err)
	}

	// Parse signer groups
	signerGroups, err := parseUint8Slice(c.String("signer-groups"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse signer-groups: %w", err)
	}
	if len(signerGroups) != len(newSigners) {
		return nil, fmt.Errorf("signer-groups length (%d) must match new-signers length (%d)", len(signerGroups), len(newSigners))
	}

	// Parse group quorums
	groupQuorums, err := parseUint8Array32(c.String("group-quorums"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse group-quorums: %w", err)
	}

	// Parse group parents
	groupParents, err := parseUint8Array32(c.String("group-parents"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse group-parents: %w", err)
	}

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
	fmt.Printf("Clear root: %v\n\n", params.clearRoot)
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

	// 1. Initialize signers instruction
	fmt.Println("1. Creating InitSigners instruction...")
	initSignersIx, err := mcmInstructions.InitSigners(mcmInstructions.InitSignersParams{
		MultisigID:   params.multisigID,
		TotalSigners: uint8(len(params.newSigners)),
		Authority:    mcmAuthority,
		ProgramID:    mcmClient.ProgramID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create init signers instruction: %w", err)
	}
	fmt.Printf("   Total signers: %d\n", len(params.newSigners))
	proposalIxs = append(proposalIxs, initSignersIx)

	// 2. Append signers instructions
	const maxSignersPerAppend = 10
	fmt.Println("\n2. Creating AppendSigners instruction(s)...")
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
		fmt.Printf("   Chunk %d: %d signers [%d:%d]\n", (i/maxSignersPerAppend)+1, len(signersChunk), i, end)
		proposalIxs = append(proposalIxs, appendSignersIx)
	}

	// 3. Finalize signers instruction
	fmt.Println("\n3. Creating FinalizeSigners instruction...")
	finalizeSignersIx, err := mcmInstructions.FinalizeSigners(mcmInstructions.FinalizeSignersParams{
		MultisigID: params.multisigID,
		Authority:  mcmAuthority,
		ProgramID:  mcmClient.ProgramID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create finalize signers instruction: %w", err)
	}
	fmt.Println("   Finalize complete")
	proposalIxs = append(proposalIxs, finalizeSignersIx)

	// 4. SetConfig instruction
	fmt.Println("\n4. Creating SetConfig instruction...")
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
	fmt.Printf("   Groups: %d, ClearRoot: %v\n", countNonZeroGroups(params.groupQuorums), params.clearRoot)
	proposalIxs = append(proposalIxs, setConfigIx)

	return proposalIxs, nil
}

// parseSigners parses comma-separated hex signer addresses
func parseSigners(signersStr string) ([][20]uint8, error) {
	signerAddrs := strings.Split(signersStr, ",")
	newSigners := make([][20]uint8, 0, len(signerAddrs))
	for _, addr := range signerAddrs {
		addr = strings.TrimSpace(addr)
		signer, err := mcmHex.Parse20(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signer %s: %w", addr, err)
		}
		newSigners = append(newSigners, signer)
	}
	return newSigners, nil
}

// parseUint8Slice parses comma-separated uint8 values into a slice
func parseUint8Slice(s string) ([]uint8, error) {
	parts := strings.Split(s, ",")
	result := make([]uint8, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var val uint8
		if _, err := fmt.Sscanf(part, "%d", &val); err != nil {
			return nil, fmt.Errorf("invalid value %s: %w", part, err)
		}
		result = append(result, val)
	}
	return result, nil
}

// parseUint8Array32 parses comma-separated uint8 values into a [32]uint8 array (zero-padded)
func parseUint8Array32(s string) ([32]uint8, error) {
	var result [32]uint8
	parts := strings.Split(s, ",")
	if len(parts) > 32 {
		return result, fmt.Errorf("too many values: got %d, max 32", len(parts))
	}
	for i, part := range parts {
		part = strings.TrimSpace(part)
		var val uint8
		if _, err := fmt.Sscanf(part, "%d", &val); err != nil {
			return result, fmt.Errorf("invalid value %s: %w", part, err)
		}
		result[i] = val
	}
	return result, nil
}

// countNonZeroGroups counts non-zero values in the group quorums array
func countNonZeroGroups(groupQuorums [32]uint8) int {
	count := 0
	for _, q := range groupQuorums {
		if q != 0 {
			count++
		}
	}
	return count
}
