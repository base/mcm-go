package signers

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// InitCommand returns the signers init command
func InitCommand() *ucli.Command {
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
			svc, err := loadSignersService(c)
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

			sig, err := svc.InitSigners(c.Context, services.InitSignersParams{
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
