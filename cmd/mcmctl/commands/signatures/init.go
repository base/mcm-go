package signatures

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// InitCommand returns the signatures init command
func InitCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Initialize signatures storage for a new root",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "root",
				Usage:    "Merkle root (32 bytes hex)",
				Required: true,
			},
			&ucli.Uint64Flag{
				Name:     "valid-until",
				Usage:    "Unix timestamp until which the root is valid",
				Required: true,
			},
			&ucli.IntFlag{
				Name:     "total",
				Usage:    "Total number of signatures",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			root, err := cli.ParseHex32(c.String("root"))
			if err != nil {
				return fmt.Errorf("invalid root: %w", err)
			}

			validUntil, err := parseValidUntil(c.Uint64("valid-until"))
			if err != nil {
				return err
			}

			total := c.Int("total")
			if total <= 0 || total > 255 {
				return fmt.Errorf("total must be between 1 and 255")
			}

			sig, err := svc.InitSignatures(c.Context, services.InitSignaturesParams{
				MultisigID:      multisigID,
				Root:            root,
				ValidUntil:      validUntil,
				TotalSignatures: uint8(total),
			})
			if err != nil {
				return fmt.Errorf("failed to init signatures: %w", err)
			}

			fmt.Printf("signatures storage initialized\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
