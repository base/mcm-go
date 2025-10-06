package proposal

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// SetRootCommand returns the proposal set-root command
func SetRootCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "set-root",
		Usage: "Load a proposal and set its root on-chain",
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

			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			proposalSvc := services.NewProposalService(mcmClient)
			sig, err := proposalSvc.SetRoot(c.Context, services.SetRootParams{
				MultisigID: pwr.MultisigID,
				Proposal:   pwr,
			})
			if err != nil {
				return fmt.Errorf("failed to set root: %w", err)
			}

			fmt.Printf("Root set successfully\n")
			fmt.Printf("  Multisig ID: 0x%x\n", pwr.MultisigID)
			fmt.Printf("  Root: 0x%x\n", pwr.Root)
			fmt.Printf("  Valid Until: %d\n", pwr.ValidUntil)
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
