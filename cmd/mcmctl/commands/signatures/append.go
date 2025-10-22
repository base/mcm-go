package signatures

import (
	"fmt"
	"strings"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/cli"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signatures append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append ECDSA signatures to storage",
		Flags: append(flags.OnchainWriteFlags(),
			flags.ProposalFlag(),
			flags.SignaturesFlag(),
		),
		Action: func(c *ucli.Context) error {
			filePath := c.String("proposal")
			pwr, err := util.LoadProposalWithRoot(filePath)
			if err != nil {
				return err
			}

			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			// Parse signatures from string to [][65]byte
			signatures, err := parseSignatures(c.String("signatures"))
			if err != nil {
				return fmt.Errorf("invalid signatures: %w", err)
			}

			// Parse program ID
			programID, err := cli.ParseProgramID(c.String("mcm-program-id"))
			if err != nil {
				return err
			}

			// Create ProposalToSign from ProposalWithRoot
			pts, err := util.CreateProposalToSign(pwr, programID)
			if err != nil {
				return fmt.Errorf("failed to create proposal to sign: %w", err)
			}

			sig, err := svc.AppendSignatures(c.Context, services.AppendSignaturesParams{
				ProposalToSign: pts,
				Signatures:     signatures,
			})
			if err != nil {
				return fmt.Errorf("failed to append signatures: %w", err)
			}

			fmt.Printf("Signatures appended successfully\n")
			fmt.Printf("Signature: %s\n", sig)
			return nil
		},
	}
}

// parseSignatures parses comma-separated ECDSA signatures to [][65]byte
func parseSignatures(s string) ([][65]byte, error) {
	if s == "" {
		return nil, fmt.Errorf("empty signatures input")
	}

	parts := strings.Split(s, ",")
	result := make([][65]byte, len(parts))

	for i, part := range parts {
		sig, err := hex.ParseSignature(part)
		if err != nil {
			return nil, fmt.Errorf("signature %d: %w", i, err)
		}
		result[i] = sig
	}

	return result, nil
}
