package proposal

import (
	"encoding/json"
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"

	ucli "github.com/urfave/cli/v2"
)

// EIP712Command returns the proposal eip712 command
func EIP712Command() *ucli.Command {
	return &ucli.Command{
		Name:  "eip712",
		Usage: "Display the complete EIP-712 typed data payload for a proposal (for external signers)",
		Flags: []ucli.Flag{
			flags.ProposalFlag(),
			flags.MCMProgramIDFlag(),
		},
		Action: func(c *ucli.Context) error {
			filePath := c.String("proposal")

			pwr, err := util.LoadProposalWithRoot(filePath)
			if err != nil {
				return err
			}

			// Parse program ID
			programID, err := util.ParseProgramID(c.String("mcm-program-id"))
			if err != nil {
				return err
			}

			// Compute message hash (which also builds the TypedData)
			pts, err := pwr.WithMessageHash(programID)
			if err != nil {
				return fmt.Errorf("failed to compute message hash: %w", err)
			}

			// Marshal the TypedData to pretty JSON
			jsonBytes, err := json.MarshalIndent(pts.TypedData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal EIP-712 data to JSON: %w", err)
			}

			fmt.Println("EIP-712 Typed Data:")
			fmt.Println("vvvvvvvv")
			fmt.Println(string(jsonBytes))
			fmt.Println("^^^^^^^^")
			return nil
		},
	}
}
