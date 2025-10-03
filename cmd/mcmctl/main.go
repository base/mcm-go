package main

import (
	"fmt"
	"os"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/services"
	"mcm-go/pkg/tx"

	ucli "github.com/urfave/cli/v2"
)

func main() {
	app := &ucli.App{
		Name:  "mcmctl",
		Usage: "CLI tool for managing MCM multisig on Solana",
		Flags: []ucli.Flag{
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
		},
		Commands: []*ucli.Command{
			{
				Name:  "multisig",
				Usage: "Multisig operations",
				Subcommands: []*ucli.Command{
					multisigInitCommand(),
				},
			},
			{
				Name:  "signers",
				Usage: "Signers management",
				Subcommands: []*ucli.Command{
					signersInitCommand(),
					signersAppendCommand(),
					signersFinalizeCommand(),
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
