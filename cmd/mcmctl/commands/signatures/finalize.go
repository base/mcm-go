package signatures

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// FinalizeCommand returns the signatures finalize command
func FinalizeCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "finalize",
		Usage: "Finalize signatures (no more additions allowed)",
		Flags: append(flags.TransactionFlags(), &ucli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "Path to proposal JSON file",
			Required: true,
		}),
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")

			pwr, err := util.LoadProposalWithRoot(filePath)
			if err != nil {
				return err
			}

			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			sig, err := svc.FinalizeSignatures(c.Context, services.FinalizeSignaturesParams{
				MultisigID: pwr.MultisigID,
				Root:       pwr.Root,
				ValidUntil: pwr.ValidUntil,
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
