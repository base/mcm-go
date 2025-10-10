package ownership

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/instructions"

	ucli "github.com/urfave/cli/v2"
)

// AcceptCommand returns the ownership accept command
func AcceptCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "accept",
		Usage: "Accept ownership of the multisig (step 2/2)",
		Flags: append(flags.OnchainWriteFlags(),
			flags.MultisigIDFlag(),
		),
		Action: func(c *ucli.Context) error {
			// Load client (authority must be the proposed_owner)
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			// Parse multisig ID
			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Build instruction
			ix, err := instructions.AcceptOwnership(instructions.AcceptOwnershipParams{
				MultisigID: multisigID,
				Authority:  mcmClient.Payer().PublicKey(),
				ProgramID:  mcmClient.ProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to build instruction: %w", err)
			}

			// Submit transaction
			sig, err := mcmClient.BuildSignAndSendWithConfirmation(c.Context, ix)
			if err != nil {
				return fmt.Errorf("failed to submit transaction: %w", err)
			}

			fmt.Printf("ownership accepted\n")
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
