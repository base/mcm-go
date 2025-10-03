package proposal

import (
	"fmt"

	"mcm-go/cmd/mcmctl/flags"
	"mcm-go/cmd/mcmctl/util"
	"mcm-go/pkg/services"

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
			&ucli.IntFlag{
				Name:    "operation-index",
				Aliases: []string{"i"},
				Usage:   "Operation index to execute (omit to execute all operations)",
				Value:   -1,
			},
		),
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")
			operationIndex := c.Int("operation-index")

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

			var operationIndices []int
			if operationIndex >= 0 {
				operationIndices = []int{operationIndex}
			} else {
				operationIndices = make([]int, len(pwr.Instructions))
				for i := range operationIndices {
					operationIndices[i] = i
				}
			}

			sigs, err := proposalSvc.Execute(c.Context, services.ExecuteParams{
				MultisigID:       pwr.MultisigID,
				ProposalWithRoot: pwr,
				OperationIndices: operationIndices,
			})
			if err != nil {
				return fmt.Errorf("failed to execute operations: %w", err)
			}

			if len(sigs) == 1 && operationIndex >= 0 {
				fmt.Printf("Operation %d executed successfully\n", operationIndex)
				fmt.Printf("signature: %s\n", sigs[0])
			} else {
				fmt.Printf("Executed %d operation(s) successfully\n", len(sigs))
				for i, sig := range sigs {
					if operationIndex >= 0 {
						fmt.Printf("  operation %d: %s\n", operationIndex, sig)
					} else {
						fmt.Printf("  operation %d: %s\n", i, sig)
					}
				}
			}

			return nil
		},
	}
}
