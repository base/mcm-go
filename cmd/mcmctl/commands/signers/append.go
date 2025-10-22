package signers

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signers append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append signers to storage",
		Flags: append(flags.OnchainWriteFlags(),
			flags.MultisigIDFlag(),
			flags.SignersFlag(),
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
			signers, sorted, err := util.ParseAndSortSigners(signersList)
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

			fmt.Printf("Signers appended successfully\n")
			fmt.Printf("Signature: %s\n", sig)
			return nil
		},
	}
}
