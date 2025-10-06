package signatures

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// ClearCommand returns the signatures clear command
func ClearCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "clear",
		Usage: "Clear signatures storage",
		Flags: append(flags.TransactionFlags(),
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
		),
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

			sig, err := svc.ClearSignatures(c.Context, services.ClearSignaturesParams{
				MultisigID: pwr.MultisigID,
				Root:       pwr.Root,
				ValidUntil: pwr.ValidUntil,
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
