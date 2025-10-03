package signers

import (
	"fmt"

	"mcm-go/cmd/mcmctl/flags"
	"mcm-go/pkg/cli"
	"mcm-go/pkg/hex"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signers append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append signers to storage",
		Flags: append(flags.TransactionFlags(),
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signers",
				Usage:    "Comma-separated list of signer addresses (must start with 0x prefix)",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			svc, err := loadSignersService(c)
			if err != nil {
				return err
			}

			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			signersList := c.String("signers")
			signers, sorted, err := cli.ParseAndSortSigners(signersList)
			if err != nil {
				return fmt.Errorf("failed to parse signers: %w", err)
			}

			if sorted {
				fmt.Println("signers were reordered to be strictly increasing")
			}

			sig, err := svc.AppendSigners(c.Context, services.AppendSignersParams{
				MultisigID:   multisigID,
				SignersBatch: signers,
			})
			if err != nil {
				return fmt.Errorf("failed to append signers: %w", err)
			}

			fmt.Printf("appended %d signers\n", len(signers))
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
