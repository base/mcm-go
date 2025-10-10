package signers

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// ClearCommand returns the signers clear command
func ClearCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "clear",
		Usage: "Clear signers storage",
		Flags: append(flags.OnchainWriteFlags(),
			flags.MultisigIDFlag(),
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

			sig, err := svc.ClearSigners(c.Context, services.ClearSignersParams{
				MultisigID: multisigID,
			})
			if err != nil {
				return fmt.Errorf("failed to clear signers: %w", err)
			}

			fmt.Printf("Signers cleared successfully: %s\n", sig)
			return nil
		},
	}
}
