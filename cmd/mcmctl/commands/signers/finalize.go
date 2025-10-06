package signers

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// FinalizeCommand returns the signers finalize command
func FinalizeCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "finalize",
		Usage: "Finalize signers (no more additions allowed)",
		Flags: append(append(flags.OnchainReadFlags(), flags.OnchainWriteFlags()...),
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
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

			sig, err := svc.FinalizeSigners(c.Context, services.FinalizeSignersParams{
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
