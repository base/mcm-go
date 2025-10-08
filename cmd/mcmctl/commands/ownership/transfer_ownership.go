package ownership

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/instructions"

	"github.com/gagliardetto/solana-go"
	ucli "github.com/urfave/cli/v2"
)

// TransferOwnershipCommand returns the multisig transfer-ownership command
func TransferOwnershipCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "transfer-ownership",
		Usage: "Propose a new owner for the multisig (step 1/2)",
		Flags: append(flags.OnchainWriteFlags(),
			flags.MultisigIDFlag(),
			flags.ProposedOwnerFlag(),
		),
		Action: func(c *ucli.Context) error {
			// Load client
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

			// Parse proposed owner
			proposedOwner, err := solana.PublicKeyFromBase58(c.String("proposed-owner"))
			if err != nil {
				return fmt.Errorf("invalid proposed-owner: %w", err)
			}

			// Build instruction
			ix, err := instructions.TransferOwnership(instructions.TransferOwnershipParams{
				MultisigID:    multisigID,
				ProposedOwner: proposedOwner,
				Authority:     mcmClient.Payer().PublicKey(),
				ProgramID:     mcmClient.ProgramID,
			})
			if err != nil {
				return fmt.Errorf("failed to build instruction: %w", err)
			}

			// Submit transaction
			sig, err := mcmClient.BuildSignAndSendWithConfirmation(c.Context, ix)
			if err != nil {
				return fmt.Errorf("failed to submit transaction: %w", err)
			}

			fmt.Printf("ownership transfer proposed\n")
			fmt.Printf("proposed owner: %s\n", proposedOwner.String())
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
