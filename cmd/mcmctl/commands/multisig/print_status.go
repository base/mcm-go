package multisig

import (
	"fmt"
	"time"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/state"

	ucli "github.com/urfave/cli/v2"
)

// PrintStatusCommand returns the multisig print-status command
func PrintStatusCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "print-status",
		Usage: "Display multisig current status (nonce, root, expiration)",
		Flags: append(flags.OnchainReadFlags(),
			flags.MultisigIDFlag(),
		),
		Action: func(c *ucli.Context) error {
			// Load client
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Fetch ExpiringRootAndOpCount
			fetcher := state.NewFetcher(mcmClient.RPC, mcmClient.ProgramID)
			rootAndOpCount, err := fetcher.GetExpiringRootAndOpCount(c.Context, multisigID)
			if err != nil {
				return fmt.Errorf("failed to fetch multisig status: %w", err)
			}

			// Display status
			fmt.Println("=== Multisig Status ===")
			fmt.Printf("Nonce (OpCount): %d\n", rootAndOpCount.OpCount)
			fmt.Printf("Root: 0x%x\n", rootAndOpCount.Root)
			fmt.Printf("Valid Until: %d", rootAndOpCount.ValidUntil)

			// Display human-readable expiration if set
			if rootAndOpCount.ValidUntil > 0 {
				expirationTime := time.Unix(int64(rootAndOpCount.ValidUntil), 0)
				fmt.Printf(" (%s)", expirationTime.UTC().Format(time.RFC3339))
			}
			fmt.Println()

			return nil
		},
	}
}
