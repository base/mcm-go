package signatures

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// FinalizeCommand returns the signatures finalize command
func FinalizeCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "finalize",
		Usage: "Finalize signatures (no more additions allowed)",
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

			sig, err := svc.FinalizeSignatures(c.Context, services.FinalizeSignaturesParams{
				MultisigID: multisigID,
				Root:       root,
				ValidUntil: validUntil,
			})
			if err != nil {
				return fmt.Errorf("failed to finalize signatures: %w", err)
			}

			fmt.Printf("signatures finalized\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
