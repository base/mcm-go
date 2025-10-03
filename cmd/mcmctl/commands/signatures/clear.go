package signatures

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// ClearCommand returns the signatures clear command
func ClearCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "clear",
		Usage: "Clear signatures storage",
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

			sig, err := svc.ClearSignatures(c.Context, services.ClearSignaturesParams{
				MultisigID: multisigID,
				Root:       root,
				ValidUntil: validUntil,
			})
			if err != nil {
				return fmt.Errorf("failed to clear signatures: %w", err)
			}

			fmt.Printf("signatures cleared\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
