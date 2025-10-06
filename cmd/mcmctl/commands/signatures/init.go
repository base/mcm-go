package signatures

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// InitCommand returns the signatures init command
func InitCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Initialize signatures storage for a new root",
		Flags: append(append(flags.OnchainReadFlags(), flags.OnchainWriteFlags()...),
			&ucli.StringFlag{
				Name: "proposal",
				// No alias to avoid conflict with --program-id
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
			&ucli.IntFlag{
				Name:     "total",
				Usage:    "Total number of signatures",
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

			total := c.Int("total")
			if total <= 0 || total > 255 {
				return fmt.Errorf("total must be between 1 and 255")
			}

			sig, err := svc.InitSignatures(c.Context, services.InitSignaturesParams{
				ProposalWithRoot: pwr,
				TotalSignatures:  uint8(total),
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
