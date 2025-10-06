package multisig

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/pda"
	"github.com/gagliardetto/solana-go"

	ucli "github.com/urfave/cli/v2"
)

// PrintAuthorityCommand returns the multisig print-authority command
func PrintAuthorityCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "print-authority",
		Usage: "Print the multisig signer PDA (authority)",
		Flags: []ucli.Flag{
			flags.ProgramIDFlag(),
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			// Parse program ID
			programID, err := solana.PublicKeyFromBase58(c.String("program-id"))
			if err != nil {
				return fmt.Errorf("invalid program ID: %w", err)
			}

			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Derive the multisig signer PDA
			signerPDA, bump, err := pda.MultisigSignerPDA(programID, multisigID)
			if err != nil {
				return fmt.Errorf("failed to derive signer PDA: %w", err)
			}

			fmt.Printf("Multisig Signer Authority (PDA): %s\n", signerPDA.String())
			fmt.Printf("Bump: %d\n", bump)
			return nil
		},
	}
}
