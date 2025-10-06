package proposal

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// ExecuteCommand returns the proposal execute command
func ExecuteCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "execute",
		Usage: "Execute operations from a proposal",
		Flags: append(flags.TransactionFlags(),
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
			&ucli.UintFlag{
				Name:    "start-index",
				Aliases: []string{"s"},
				Usage:   "Index of first operation to execute",
				Value:   0,
			},
			&ucli.UintFlag{
				Name:     "operation-count",
				Aliases:  []string{"n"},
				Usage:    "Number of operations to execute",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")
			startIndex := c.Uint("start-index")
			operationCount := c.Uint("operation-count")

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

			sigs, err := proposalSvc.Execute(c.Context, services.ExecuteParams{
				MultisigID:       pwr.MultisigID,
				ProposalWithRoot: pwr,
				StartIndex:       startIndex,
				OperationCount:   operationCount,
			})
			if err != nil {
				return fmt.Errorf("failed to execute operations: %w", err)
			}

			fmt.Printf("Executed %d operation(s) successfully\n", len(sigs))
			for i, sig := range sigs {
				opIdx := startIndex + uint(i)
				fmt.Printf("  operation %d: %s\n", opIdx, sig)
			}

			return nil
		},
	}
}
