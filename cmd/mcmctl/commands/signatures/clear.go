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
		Flags: append(append(flags.OnchainReadFlags(), flags.OnchainWriteFlags()...),
			&ucli.StringFlag{
				Name: "proposal",
				// No alias to avoid conflict with --program-id
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			filePath := c.String("proposal")

			pwr, err := util.LoadProposalWithRoot(filePath)
			if err != nil {
				return err
			}

			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			sig, err := svc.ClearSignatures(c.Context, services.ClearSignaturesParams{
				ProposalWithRoot: pwr,
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
