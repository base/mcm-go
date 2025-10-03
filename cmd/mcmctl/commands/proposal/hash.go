package proposal

import (
	"encoding/hex"
	"fmt"

	proposalIO "mcm-go/pkg/proposal/io"

	ucli "github.com/urfave/cli/v2"
)

// HashCommand returns the proposal hash command
func HashCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "hash",
		Usage: "Compute the hash to sign for a proposal (offline operation)",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")

			// Load proposal from file
			p, err := proposalIO.LoadProposal(filePath)
			if err != nil {
				return fmt.Errorf("failed to load proposal: %w", err)
			}

			// Compute root and proofs
			pwr, err := p.WithRoot()
			if err != nil {
				return fmt.Errorf("failed to compute Merkle root: %w", err)
			}

			// Compute hash to sign
			pts, err := pwr.WithHashToSign()
			if err != nil {
				return fmt.Errorf("failed to compute hash to sign: %w", err)
			}

			// Display proposal info
			fmt.Printf("Proposal loaded successfully\n")
			fmt.Printf("  Multisig ID: 0x%x\n", pts.MultisigID)
			fmt.Printf("  Valid Until: %d\n", pts.ValidUntil)
			fmt.Printf("  Instructions: %d\n", len(pts.Instructions))
			fmt.Printf("  Chain ID: %d\n", pts.RootMetadata.ChainID)
			fmt.Printf("  Pre Op Count: %d\n", pts.RootMetadata.PreOpCount)
			fmt.Printf("  Post Op Count: %d\n", pts.RootMetadata.PostOpCount)
			fmt.Printf("  Override Previous Root: %v\n", pts.RootMetadata.OverridePreviousRoot)
			fmt.Printf("\n")
			fmt.Printf("Merkle Root: 0x%x\n", pts.Root)
			fmt.Printf("\n")
			fmt.Printf("Hash to Sign (keccak256(root || validUntil)):\n")
			fmt.Print("vvvvvvvv\n")
			fmt.Printf("0x%s\n", hex.EncodeToString(pts.HashToSign[:]))
			fmt.Println("^^^^^^^^")

			return nil
		},
	}
}
