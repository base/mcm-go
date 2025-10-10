package concrete

import (
	"context"
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/cli"
	"github.com/base/mcm-go/pkg/hex"
	proposalIO "github.com/base/mcm-go/pkg/proposal/io"
	"github.com/base/mcm-go/pkg/services"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	ucli "github.com/urfave/cli/v2"
)

// UpgradeCommand returns the upgrade proposal creation command
func UpgradeCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "create-upgrade",
		Usage: "Create a proposal to upgrade a Solana program using BPF Loader v3",
		Flags: append(flags.OnchainReadFlags(),
			flags.MultisigIDFlag(),
			flags.ValidUntilFlag(),
			flags.OverridePreviousRootFlag(),
			flags.OutputFlag(),
			&ucli.StringFlag{
				Name:     "program",
				Usage:    "Program account address to upgrade",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "buffer",
				Usage:    "Buffer account address with new program data",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "spill",
				Usage:    "Spill account address to receive refunded lamports",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			// Parse and validate CLI parameters
			params, err := parseUpgradeParams(c)
			if err != nil {
				return err
			}

			// Fetch and validate authorities from on-chain
			upgradeAuthority, err := fetchAndValidateAuthorities(params)
			if err != nil {
				return err
			}

			// Create upgrade instruction
			upgradeIx := newUpgradeInstruction(
				params.programData,
				params.program,
				params.buffer,
				params.spill,
				upgradeAuthority,
			)

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
				ValidUntil:           uint32(c.Uint64("valid-until")),
				Instructions:         []solana.Instruction{upgradeIx},
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

			fmt.Printf("\nUpgrade proposal created successfully and saved to %s\n", outputPath)
			return nil
		},
	}
}

// upgradeParams holds parsed parameters for the upgrade command
type upgradeParams struct {
	program     solana.PublicKey
	buffer      solana.PublicKey
	spill       solana.PublicKey
	programData solana.PublicKey
	multisigID  [32]byte
	rpcURL      string
}

// parseUpgradeParams parses and validates CLI parameters for upgrade command
func parseUpgradeParams(c *ucli.Context) (*upgradeParams, error) {
	program, err := solana.PublicKeyFromBase58(c.String("program"))
	if err != nil {
		return nil, fmt.Errorf("invalid program address: %w", err)
	}

	buffer, err := solana.PublicKeyFromBase58(c.String("buffer"))
	if err != nil {
		return nil, fmt.Errorf("invalid buffer address: %w", err)
	}

	spill, err := solana.PublicKeyFromBase58(c.String("spill"))
	if err != nil {
		return nil, fmt.Errorf("invalid spill address: %w", err)
	}

	multisigID, err := hex.Parse32(c.String("multisig-id"))
	if err != nil {
		return nil, fmt.Errorf("invalid multisig-id: %w", err)
	}

	rpcURL := cli.ResolveNetworkAlias(c.String("rpc-url"), false)

	// Derive ProgramData PDA from program address
	programData, _, err := solana.FindProgramAddress(
		[][]byte{program.Bytes()},
		solana.BPFLoaderUpgradeableProgramID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive program data address: %w", err)
	}
	fmt.Printf("ProgramData (derived): %s\n", programData)

	return &upgradeParams{
		program:     program,
		buffer:      buffer,
		spill:       spill,
		programData: programData,
		multisigID:  multisigID,
		rpcURL:      rpcURL,
	}, nil
}

// fetchAndValidateAuthorities fetches authorities from on-chain and validates they match
func fetchAndValidateAuthorities(params *upgradeParams) (solana.PublicKey, error) {
	ctx := context.Background()
	rpcClient := rpc.New(params.rpcURL)

	upgradeAuthority, err := fetchUpgradeAuthority(ctx, rpcClient, params.program, params.programData)
	if err != nil {
		return solana.PublicKey{}, err
	}
	fmt.Printf("Upgrade authority: %s\n", upgradeAuthority)

	bufferAuthority, err := fetchBufferAuthority(ctx, rpcClient, params.buffer)
	if err != nil {
		return solana.PublicKey{}, err
	}
	fmt.Printf("Buffer authority: %s\n", bufferAuthority)

	// Validate that authorities match
	if upgradeAuthority != bufferAuthority {
		return solana.PublicKey{}, fmt.Errorf("program authority (%s) does not match buffer authority (%s)", upgradeAuthority, bufferAuthority)
	}

	return upgradeAuthority, nil
}

// fetchUpgradeAuthority fetches and parses the upgrade authority from the program data account
func fetchUpgradeAuthority(ctx context.Context, client *rpc.Client, program, programData solana.PublicKey) (solana.PublicKey, error) {
	// Fetch program account
	programAccountInfo, err := client.GetAccountInfo(ctx, program)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to fetch program account: %w", err)
	}
	if programAccountInfo == nil || programAccountInfo.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("program account not found")
	}

	// Fetch program data account
	programDataAccountInfo, err := client.GetAccountInfo(ctx, programData)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to fetch program data account: %w", err)
	}
	if programDataAccountInfo == nil || programDataAccountInfo.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("program data account not found")
	}

	// Parse upgrade authority from program data account
	// ProgramData account layout: [discriminator(4), slot(8), upgrade_authority(32), ...]
	programDataBytes := programDataAccountInfo.Value.Data.GetBinary()
	if len(programDataBytes) < 45 {
		return solana.PublicKey{}, fmt.Errorf("invalid program data account size: expected at least 45 bytes, got %d", len(programDataBytes))
	}

	// Check if authority is present (option discriminator at byte 12)
	hasAuthority := programDataBytes[12] == 1
	if !hasAuthority {
		return solana.PublicKey{}, fmt.Errorf("program has no upgrade authority")
	}

	// Extract authority pubkey (bytes 13-45)
	var upgradeAuthority solana.PublicKey
	copy(upgradeAuthority[:], programDataBytes[13:45])

	return upgradeAuthority, nil
}

// fetchBufferAuthority fetches and parses the buffer authority from the buffer account
func fetchBufferAuthority(ctx context.Context, client *rpc.Client, buffer solana.PublicKey) (solana.PublicKey, error) {
	// Fetch buffer account
	bufferAccountInfo, err := client.GetAccountInfo(ctx, buffer)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to fetch buffer account: %w", err)
	}
	if bufferAccountInfo == nil || bufferAccountInfo.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("buffer account not found")
	}

	// Parse buffer authority
	// Buffer account layout: [discriminator(4), authority(32), ...]
	bufferBytes := bufferAccountInfo.Value.Data.GetBinary()
	if len(bufferBytes) < 37 {
		return solana.PublicKey{}, fmt.Errorf("invalid buffer account size: expected at least 37 bytes, got %d", len(bufferBytes))
	}

	// Check if buffer authority is present (option discriminator at byte 4)
	hasBufferAuthority := bufferBytes[4] == 1
	if !hasBufferAuthority {
		return solana.PublicKey{}, fmt.Errorf("buffer has no authority")
	}

	// Extract buffer authority pubkey (bytes 5-37)
	var bufferAuthority solana.PublicKey
	copy(bufferAuthority[:], bufferBytes[5:37])

	return bufferAuthority, nil
}

// newUpgradeInstruction builds an "upgrade" instruction for the BPF Loader Upgradeable program.
// The data (0x03000000) is the 4-byte upgrade instruction discriminator.
func newUpgradeInstruction(
	programData solana.PublicKey,
	program solana.PublicKey,
	buffer solana.PublicKey,
	spill solana.PublicKey,
	authority solana.PublicKey,
) solana.Instruction {
	accounts := solana.AccountMetaSlice{}

	// Account 0: ProgramData account (writable)
	accounts.Append(solana.NewAccountMeta(programData, true, false))
	// Account 1: Program account (writable)
	accounts.Append(solana.NewAccountMeta(program, true, false))
	// Account 2: Buffer account with new program data (writable)
	accounts.Append(solana.NewAccountMeta(buffer, true, false))
	// Account 3: Spill account (writable)
	accounts.Append(solana.NewAccountMeta(spill, true, false))
	// Account 4: Rent sysvar (read-only)
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	// Account 5: Clock sysvar (read-only)
	accounts.Append(solana.NewAccountMeta(solana.SysVarClockPubkey, false, false))
	// Account 6: Upgrade authority (signer)
	accounts.Append(solana.NewAccountMeta(authority, false, true))

	// Create the instruction with upgrade discriminator (0x03000000)
	return solana.NewInstruction(
		solana.BPFLoaderUpgradeableProgramID,
		accounts,
		[]byte{0x03, 0x00, 0x00, 0x00},
	)
}
