package loader_v3

import (
	"context"
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	ucli "github.com/urfave/cli/v2"
)

// SetAuthorityCommand returns the set authority proposal creation command
func SetAuthorityCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "set-authority",
		Usage: "Create a proposal to change or remove the upgrade authority of a BPF Loader v3 program or buffer",
		Flags: append(flags.ProposalCreationFlags(),
			&ucli.StringFlag{
				Name:     "account",
				Usage:    "Program ID or Buffer account address",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "new-authority",
				Usage:    "New authority address (omit to make the program/buffer immutable)",
				Required: false,
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parseSetAuthorityParams(c)
			if err != nil {
				return err
			}

			// Detect account type and resolve target address
			target, currentAuthority, accountType, err := detectAndResolveTarget(params)
			if err != nil {
				return err
			}

			fmt.Printf("Account type: %s\n", accountType)
			fmt.Printf("Target account: %s\n", target)
			fmt.Printf("Current authority: %s\n", currentAuthority)

			// Warn user if making immutable
			if params.newAuthority == nil {
				fmt.Println("WARNING: This will make the program/buffer IMMUTABLE.")
				fmt.Println("This action is IRREVERSIBLE - no further upgrades or authority changes will be possible.")
			} else {
				fmt.Printf("New authority: %s\n", *params.newAuthority)
			}

			// Create set authority instruction
			setAuthorityIx := newSetAuthorityInstruction(
				target,
				currentAuthority,
				params.newAuthority,
			)

			// Load MCM client and create proposal service
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			proposalSvc := services.NewProposalService(mcmClient)

			// Create proposal by fetching on-chain state
			fmt.Println("\nFetching MCM on-chain state...")
			p, err := proposalSvc.CreateProposalFromChain(c.Context, services.CreateProposalFromChainParams{
				MultisigID:           params.multisigID,
				ValidUntil:           params.validUntil,
				Instructions:         []solana.Instruction{setAuthorityIx},
				OverridePreviousRoot: params.overridePreviousRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to create proposal: %w", err)
			}

			// Save proposal to file
			if err := proposalIO.SaveProposal(p, params.outputPath); err != nil {
				return fmt.Errorf("failed to save proposal: %w", err)
			}

			action := "change authority"
			if params.newAuthority == nil {
				action = "make immutable"
			}
			fmt.Printf("\nSet authority proposal (%s) created successfully and saved to %s\n", action, params.outputPath)
			return nil
		},
	}
}

// setAuthorityParams holds parsed parameters for the set authority command
type setAuthorityParams struct {
	// Common MCM parameters
	multisigID           [32]byte
	validUntil           uint32
	overridePreviousRoot bool
	outputPath           string
	rpcURL               string
	// Specific loader_v3 parameters
	account      solana.PublicKey
	newAuthority *solana.PublicKey // nil means make immutable
}

// parseSetAuthorityParams parses and validates CLI parameters for set authority command
func parseSetAuthorityParams(c *ucli.Context) (*setAuthorityParams, error) {
	// 1. Parse common MCM parameters
	multisigID, err := hex.Parse32(c.String("multisig-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid multisig-id: %w", err)
	}

	validUntil := uint32(c.Uint64("valid-until"))
	overridePreviousRoot := c.Bool("override-previous-root")
	outputPath := c.String("output")
	rpcURL := util.ResolveNetworkAlias(c.String("rpc-url"), false)

	// 2. Parse specific loader_v3 parameters
	account, err := solana.PublicKeyFromBase58(c.String("account"))
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %w", err)
	}

	// Parse new authority (optional)
	var newAuthority *solana.PublicKey
	if c.IsSet("new-authority") {
		newAuth, err := solana.PublicKeyFromBase58(c.String("new-authority"))
		if err != nil {
			return nil, fmt.Errorf("invalid new-authority address: %w", err)
		}
		newAuthority = &newAuth
	}

	// 3. Return with common fields first
	return &setAuthorityParams{
		multisigID:           multisigID,
		validUntil:           validUntil,
		overridePreviousRoot: overridePreviousRoot,
		outputPath:           outputPath,
		rpcURL:               rpcURL,
		account:              account,
		newAuthority:         newAuthority,
	}, nil
}

// detectAndResolveTarget detects if the account is a Program or Buffer and resolves the target address
// Returns: target address, current authority, account type description, error
func detectAndResolveTarget(params *setAuthorityParams) (solana.PublicKey, solana.PublicKey, string, error) {
	ctx := context.Background()
	rpcClient := rpc.New(params.rpcURL)

	// Fetch the account to determine its type
	accountInfo, err := rpcClient.GetAccountInfo(ctx, params.account)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, "", fmt.Errorf("failed to fetch account: %w", err)
	}
	if accountInfo == nil || accountInfo.Value == nil {
		return solana.PublicKey{}, solana.PublicKey{}, "", fmt.Errorf("account not found: %s", params.account)
	}

	// Check if the account is owned by BPF Loader Upgradeable
	if accountInfo.Value.Owner != solana.BPFLoaderUpgradeableProgramID {
		return solana.PublicKey{}, solana.PublicKey{}, "", fmt.Errorf("account is not owned by BPF Loader Upgradeable program (owner: %s)", accountInfo.Value.Owner)
	}

	// Detect account type based on executable flag
	if accountInfo.Value.Executable {
		// This is a Program executable - derive ProgramData PDA
		programData, _, err := solana.FindProgramAddress(
			[][]byte{params.account.Bytes()},
			solana.BPFLoaderUpgradeableProgramID,
		)
		if err != nil {
			return solana.PublicKey{}, solana.PublicKey{}, "", fmt.Errorf("failed to derive program data address: %w", err)
		}

		fmt.Printf("Program ID detected: %s\n", params.account)
		fmt.Printf("ProgramData (derived): %s\n", programData)

		// Fetch upgrade authority from ProgramData
		upgradeAuthority, err := fetchUpgradeAuthority(ctx, rpcClient, params.account, programData)
		if err != nil {
			return solana.PublicKey{}, solana.PublicKey{}, "", err
		}

		return programData, upgradeAuthority, "Program", nil
	} else {
		// This is a Buffer (not executable)
		fmt.Printf("Buffer detected: %s\n", params.account)

		bufferAuthority, err := fetchBufferAuthority(ctx, rpcClient, params.account)
		if err != nil {
			return solana.PublicKey{}, solana.PublicKey{}, "", err
		}

		return params.account, bufferAuthority, "Buffer", nil
	}
}

// newSetAuthorityInstruction builds a "set_authority" instruction for the BPF Loader Upgradeable program.
// The data (0x04000000) is the 4-byte set authority instruction discriminator.
// TODO: This should probably be moved to some internal/loader_v3 package
func newSetAuthorityInstruction(
	target solana.PublicKey, // Buffer or ProgramData account
	currentAuthority solana.PublicKey, // Current authority (signer)
	newAuthority *solana.PublicKey, // New authority (nil = make immutable)
) solana.Instruction {
	accounts := solana.AccountMetaSlice{}

	// Account 0: Buffer or ProgramData account (writable)
	accounts.Append(solana.NewAccountMeta(target, true, false))
	// Account 1: Current authority (signer)
	accounts.Append(solana.NewAccountMeta(currentAuthority, false, true))
	// Account 2: New authority (optional - omit to make immutable)
	if newAuthority != nil {
		accounts.Append(solana.NewAccountMeta(*newAuthority, false, false))
	}

	// Create the instruction with set authority discriminator (0x04000000)
	return solana.NewInstruction(
		solana.BPFLoaderUpgradeableProgramID,
		accounts,
		[]byte{0x04, 0x00, 0x00, 0x00},
	)
}
