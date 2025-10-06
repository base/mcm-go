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
		Flags: append(append(flags.OnchainReadFlags(), flags.OnchainWriteFlags()...), &ucli.StringFlag{
			Name: "proposal",
			// No alias to avoid conflict with --program-id
			Usage:    "Path to proposal JSON file",
			Required: true,
		}),
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

			sig, err := svc.FinalizeSignatures(c.Context, services.FinalizeSignaturesParams{
				ProposalWithRoot: pwr,
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
