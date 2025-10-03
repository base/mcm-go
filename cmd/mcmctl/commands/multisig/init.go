package multisig

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/tx"

	ucli "github.com/urfave/cli/v2"
)

// InitCommand returns the multisig init command
func InitCommand() *ucli.Command {
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
			// Load client
			mcmClient, err := loadClient(c)
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
				Authority:  mcmClient.Payer.PublicKey(),
				ProgramID:  mcmClient.ProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to build instruction: %w", err)
			}

			// Submit transaction
			sig, err := tx.NewTxBuilder(mcmClient.RPC, mcmClient.WS, mcmClient.Payer).
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

// loadClient loads the MCM client from CLI flags
func loadClient(c *ucli.Context) (*client.Client, error) {
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc"),
		WSUrl:       c.String("ws"),
		ProgramID:   c.String("program-id"),
		KeypairPath: c.String("authority"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return client.New(*cfg)
}
