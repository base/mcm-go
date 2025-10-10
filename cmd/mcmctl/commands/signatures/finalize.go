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
		Flags: append(flags.OnchainWriteFlags(),
			flags.ProposalFlag(),
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

			sig, err := svc.FinalizeSignatures(c.Context, services.FinalizeSignaturesParams{
				ProposalWithRoot: pwr,
			})
			if err != nil {
				return fmt.Errorf("failed to finalize signatures: %w", err)
			}

			fmt.Printf("Signatures finalized successfully: %s\n", sig)
			return nil
		},
	}
}
