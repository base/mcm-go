package signers

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signers append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append signers to storage",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signers",
				Usage:    "Comma-separated list of signer addresses (hex, with/without 0x prefix)",
				Required: true,
			},
			&ucli.IntFlag{
				Name:  "batch-size",
				Usage: "Batch size per transaction (1-32)",
				Value: 10,
			},
		},
		Action: func(c *ucli.Context) error {
			svc, err := loadSignersService(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			batchSize := c.Int("batch-size")
			if batchSize < 1 || batchSize > 32 {
				return fmt.Errorf("batch-size must be between 1 and 32")
			}

			// Parse and validate signers
			signersList := c.String("signers")
			signers, sorted, err := cli.ParseAndSortSigners(signersList)
			if err != nil {
				return fmt.Errorf("failed to parse signers: %w", err)
			}

			if sorted {
				fmt.Println("signers were reordered to be strictly increasing")
			}

			sigs, err := svc.AppendSignersInBatches(c.Context, services.AppendSignersInBatchesParams{
				MultisigID: multisigID,
				Signers:    signers,
				BatchSize:  batchSize,
			})
			if err != nil {
				return fmt.Errorf("failed to append signers: %w", err)
			}

			fmt.Printf("appended %d signers in %d batch(es)\n", len(signers), len(sigs))
			fmt.Printf("final signature: %s\n", sigs[len(sigs)-1])
			return nil
		},
	}
}
