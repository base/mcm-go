package multisig

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/instructions"
	"github.com/base/mcm-go/pkg/tx"

	ucli "github.com/urfave/cli/v2"
)

// InitCommand returns the multisig init command
func InitCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Initialize a new multisig",
		Flags: append(flags.TransactionFlags(),
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
		),
		Action: func(c *ucli.Context) error {
			// Load client
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			multisigID, err := hex.Parse32(c.String("multisig-id"))
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
