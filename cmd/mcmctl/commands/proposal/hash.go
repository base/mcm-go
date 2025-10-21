package proposal

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"

	ucli "github.com/urfave/cli/v2"
)

// HashCommand returns the proposal hash command
func HashCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "hash",
		Usage: "Compute the hash to sign for a proposal (offline operation)",
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

			pts, err := pwr.WithHashToSign(programID)
			if err != nil {
				return fmt.Errorf("failed to compute hash to sign: %w", err)
			}

			// Display proposal info
			fmt.Printf("Merkle Root: 0x%x\n", pts.Root)
			fmt.Println("Hash to Sign (EIP-712):")
			fmt.Println("vvvvvvvv")
			fmt.Printf("0x1901%x%x\n", pts.DomainSeparator, pts.StructHash)
			fmt.Println("^^^^^^^^")

			return nil
		},
	}
}
