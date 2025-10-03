package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	proposalIO "mcm-go/pkg/proposal/io"
	"mcm-go/pkg/services"
	"mcm-go/pkg/tx"

	ucli "github.com/urfave/cli/v2"
)

// blockchainFlags are required for commands that interact with Solana blockchain
var blockchainFlags = []ucli.Flag{
	&ucli.StringFlag{
		Name:     "rpc",
		Aliases:  []string{"r"},
		Usage:    "Solana RPC endpoint URL",
		EnvVars:  []string{"MCM_RPC_URL"},
		Required: true,
	},
	&ucli.StringFlag{
		Name:     "ws",
		Usage:    "Solana WebSocket endpoint URL (required for confirmations)",
		EnvVars:  []string{"MCM_WS_URL"},
		Required: true,
	},
	&ucli.StringFlag{
		Name:     "program-id",
		Aliases:  []string{"p"},
		Usage:    "MCM program ID (base58)",
		EnvVars:  []string{"MCM_PROGRAM_ID"},
		Required: true,
	},
	&ucli.StringFlag{
		Name:    "authority",
		Aliases: []string{"a"},
		Usage:   "Path to authority keypair file (JSON or base58, also used as transaction payer)",
		EnvVars: []string{"MCM_AUTHORITY"},
		Value:   cli.DefaultKeypairPath(),
	},
}

func main() {
	app := &ucli.App{
		Name:  "mcmctl",
		Usage: "CLI tool for managing MCM multisig on Solana",
		Commands: []*ucli.Command{
			{
				Name:  "multisig",
				Usage: "Multisig operations",
				Flags: blockchainFlags,
				Subcommands: []*ucli.Command{
					multisigInitCommand(),
				},
			},
			{
				Name:  "signers",
				Usage: "Signers management",
				Flags: blockchainFlags,
				Subcommands: []*ucli.Command{
					signersInitCommand(),
					signersAppendCommand(),
					signersFinalizeCommand(),
					signersSetConfigCommand(),
				},
			},
			{
				Name:  "proposal",
				Usage: "Proposal operations (offline and online)",
				Subcommands: []*ucli.Command{
					proposalSignCommand(),
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type AppContext struct {
	Client          *client.Client
	SignersService  *services.SignersService
	ProposalService *services.ProposalService
}

// getAppContext initializes and returns the app context from global flags
func getAppContext(c *ucli.Context) (*AppContext, error) {
	// Load config and create client
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc"),
		WSUrl:       c.String("ws"),
		ProgramID:   c.String("program-id"),
		KeypairPath: c.String("authority"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	mcmClient, err := client.New(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &AppContext{
		Client:          mcmClient,
		SignersService:  services.NewSignersService(mcmClient),
		ProposalService: services.NewProposalService(mcmClient),
	}, nil
}

func multisigInitCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Initialize a new multisig",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.Uint64Flag{
				Name:     "chain-id",
				Usage:    "Chain ID (uint64)",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			// Initialize context
			ctx, err := getAppContext(c)
			if err != nil {
				return err
			}

			// Parse multisig ID
			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Build instruction
			ix, err := instructions.Initialize(instructions.InitializeParams{
				ChainID:    c.Uint64("chain-id"),
				MultisigID: multisigID,
				Authority:  ctx.Client.Payer.PublicKey(),
				ProgramID:  ctx.Client.ProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to build instruction: %w", err)
			}

			// Submit transaction
			sig, err := tx.NewTxBuilder(ctx.Client.RPC, ctx.Client.WS, ctx.Client.Payer).
				AddInstruction(ix).
				BuildSignAndSendWithConfirmation(c.Context)
			if err != nil {
				return fmt.Errorf("failed to submit transaction: %w", err)
			}

			fmt.Printf("multisig initialized\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

func signersInitCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Initialize signers storage",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.IntFlag{
				Name:     "total",
				Usage:    "Total number of signers",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			ctx, err := getAppContext(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			total := c.Int("total")
			if total <= 0 || total > 255 {
				return fmt.Errorf("total must be between 1 and 255")
			}

			sig, err := ctx.SignersService.InitSigners(c.Context, services.InitSignersParams{
				MultisigID:   multisigID,
				TotalSigners: uint8(total),
			})
			if err != nil {
				return fmt.Errorf("failed to init signers: %w", err)
			}

			fmt.Printf("signers storage initialized\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

func signersAppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append signers to storage",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signers",
				Usage:    "Comma-separated list of signer addresses (hex, with/without 0x prefix)",
				Required: true,
			},
			&ucli.IntFlag{
				Name:  "batch-size",
				Usage: "Batch size per transaction (1-32)",
				Value: 10,
			},
		},
		Action: func(c *ucli.Context) error {
			ctx, err := getAppContext(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			batchSize := c.Int("batch-size")
			if batchSize < 1 || batchSize > 32 {
				return fmt.Errorf("batch-size must be between 1 and 32")
			}

			// Parse and validate signers
			signersList := c.String("signers")
			signers, sorted, err := cli.ParseAndSortSigners(signersList)
			if err != nil {
				return fmt.Errorf("failed to parse signers: %w", err)
			}

			if sorted {
				fmt.Println("signers were reordered to be strictly increasing")
			}

			sigs, err := ctx.SignersService.AppendSignersInBatches(c.Context, services.AppendSignersInBatchesParams{
				MultisigID: multisigID,
				Signers:    signers,
				BatchSize:  batchSize,
			})
			if err != nil {
				return fmt.Errorf("failed to append signers: %w", err)
			}

			fmt.Printf("appended %d signers in %d batch(es)\n", len(signers), len(sigs))
			fmt.Printf("final signature: %s\n", sigs[len(sigs)-1])
			return nil
		},
	}
}

func signersFinalizeCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "finalize",
		Usage: "Finalize signers (no more additions allowed)",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			ctx, err := getAppContext(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			sig, err := ctx.SignersService.FinalizeSigners(c.Context, services.FinalizeSignersParams{
				MultisigID: multisigID,
			})
			if err != nil {
				return fmt.Errorf("failed to finalize signers: %w", err)
			}

			fmt.Printf("signers finalized\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

func signersSetConfigCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "set-config",
		Usage: "Set the multisig configuration (signer groups and quorums)",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signer-groups",
				Usage:    "Comma-separated group assignment for each signer (e.g., '0,1,0' for 3 signers)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "group-quorums",
				Usage:    "Comma-separated quorum for each group (automatically padded to 32, e.g., '1' or '2,1,1')",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "group-parents",
				Usage:    "Comma-separated parent group for each group (automatically padded to 32, e.g., '0' or '0,0,1')",
				Required: true,
			},
			&ucli.BoolFlag{
				Name:  "clear-root",
				Usage: "Clear the current Merkle root (invalidates pending operations)",
				Value: false,
			},
		},
		Action: func(c *ucli.Context) error {
			ctx, err := getAppContext(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Parse signer groups
			signerGroups, err := parseUint8Slice(c.String("signer-groups"))
			if err != nil {
				return fmt.Errorf("invalid signer-groups: %w", err)
			}

			// Parse and pad group quorums to 32 elements
			groupQuorums, err := parseAndPadUint8Array32(c.String("group-quorums"))
			if err != nil {
				return fmt.Errorf("invalid group-quorums: %w", err)
			}

			// Parse and pad group parents to 32 elements
			groupParents, err := parseAndPadUint8Array32(c.String("group-parents"))
			if err != nil {
				return fmt.Errorf("invalid group-parents: %w", err)
			}

			clearRoot := c.Bool("clear-root")

			sig, err := ctx.SignersService.SetConfig(c.Context, services.SetConfigParams{
				MultisigID:   multisigID,
				SignerGroups: signerGroups,
				GroupQuorums: groupQuorums,
				GroupParents: groupParents,
				ClearRoot:    clearRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}

			fmt.Printf("configuration set successfully\n")
			fmt.Printf("  signer groups: %v\n", signerGroups)
			fmt.Printf("  group quorums: %v (first %d)\n", groupQuorums[:min(5, len(groupQuorums))], min(5, len(groupQuorums)))
			fmt.Printf("  group parents: %v (first %d)\n", groupParents[:min(5, len(groupParents))], min(5, len(groupParents)))
			fmt.Printf("  clear root: %v\n", clearRoot)
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

// parseUint8Slice parses a comma-separated string into []uint8
func parseUint8Slice(s string) ([]byte, error) {
	if s == "" {
		return nil, fmt.Errorf("empty input")
	}

	parts := strings.Split(s, ",")
	result := make([]byte, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid value at index %d: %s", i, part)
		}
		result[i] = uint8(val)
	}

	return result, nil
}

// parseAndPadUint8Array32 parses a comma-separated string and pads it to exactly 32 elements with zeros
func parseAndPadUint8Array32(s string) ([32]uint8, error) {
	var result [32]uint8

	if s == "" {
		return result, fmt.Errorf("empty input")
	}

	parts := strings.Split(s, ",")
	if len(parts) > 32 {
		return result, fmt.Errorf("too many values: got %d, max 32", len(parts))
	}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return result, fmt.Errorf("invalid value at index %d: %s", i, part)
		}
		result[i] = uint8(val)
	}

	// Remaining elements are already 0 due to zero-initialization
	return result, nil
}

func proposalSignCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "sign",
		Usage: "Load a proposal and display the hash to sign (offline operation)",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")

			// Load proposal from file
			p, err := proposalIO.LoadProposal(filePath)
			if err != nil {
				return fmt.Errorf("failed to load proposal: %w", err)
			}

			// Compute root and proofs
			pwr, err := p.WithRoot()
			if err != nil {
				return fmt.Errorf("failed to compute Merkle root: %w", err)
			}

			// Compute hash to sign
			pts, err := pwr.WithHashToSign()
			if err != nil {
				return fmt.Errorf("failed to compute hash to sign: %w", err)
			}

			// Display proposal info
			fmt.Printf("Proposal loaded successfully\n")
			fmt.Printf("  Multisig ID: 0x%x\n", pts.MultisigID)
			fmt.Printf("  Valid Until: %d\n", pts.ValidUntil)
			fmt.Printf("  Instructions: %d\n", len(pts.Instructions))
			fmt.Printf("  Chain ID: %d\n", pts.RootMetadata.ChainID)
			fmt.Printf("  Pre Op Count: %d\n", pts.RootMetadata.PreOpCount)
			fmt.Printf("  Post Op Count: %d\n", pts.RootMetadata.PostOpCount)
			fmt.Printf("  Override Previous Root: %v\n", pts.RootMetadata.OverridePreviousRoot)
			fmt.Printf("\n")
			fmt.Printf("Merkle Root: 0x%x\n", pts.Root)
			fmt.Printf("\n")
			fmt.Printf("Hash to Sign (keccak256(root || validUntil)):\n")
			fmt.Printf("vvvvvvvv")
			fmt.Printf("0x%s\n", hex.EncodeToString(pts.HashToSign[:]))
			fmt.Printf("^^^^^^^^")

			return nil
		},
	}
}
