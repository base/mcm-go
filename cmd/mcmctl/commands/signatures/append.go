package signatures

import (
	"fmt"
	"strconv"
	"strings"

	"mcm-go/cmd/mcmctl/flags"
	"mcm-go/cmd/mcmctl/util"
	"mcm-go/pkg/bindings"
	"mcm-go/pkg/hex"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signatures append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append ECDSA signatures to storage",
		Flags: append(flags.TransactionFlags(),
			&ucli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to proposal JSON file",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signatures",
				Usage:    "Comma-separated ECDSA signatures (must start with 0x, format: 0x<130 hex chars>)",
				Required: true,
			},
		),
		Action: func(c *ucli.Context) error {
			filePath := c.String("file")
			pwr, err := util.LoadProposalWithRoot(filePath)
			if err != nil {
				return err
			}

			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			sigsBatch, err := parseSignatures(c.String("signatures"))
			if err != nil {
				return fmt.Errorf("invalid signatures: %w", err)
			}

			sig, err := svc.AppendSignatures(c.Context, services.AppendSignaturesParams{
				MultisigID:      pwr.MultisigID,
				Root:            pwr.Root,
				ValidUntil:      pwr.ValidUntil,
				SignaturesBatch: sigsBatch,
			})
			if err != nil {
				return fmt.Errorf("failed to append signatures: %w", err)
			}

			fmt.Printf("appended %d signature(s)\n", len(sigsBatch))
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}

func parseSignatures(s string) ([]bindings.Signature, error) {
	if s == "" {
		return nil, fmt.Errorf("empty signatures input")
	}

	parts := strings.Split(s, ",")
	result := make([]bindings.Signature, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if !strings.HasPrefix(part, "0x") {
			return nil, fmt.Errorf("signature %d: must start with '0x' prefix", i)
		}

		if len(part) != 132 { // 0x + 130 hex chars
			return nil, fmt.Errorf("signature %d: expected 0x followed by 130 hex chars (65 bytes), got %d total chars", i, len(part))
		}

		// Extract r, s, v with 0x prefix for parsing
		rBytes, err := hex.Parse32("0x" + part[2:66])
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid r value: %w", i, err)
		}

		sBytes, err := hex.Parse32("0x" + part[66:130])
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid s value: %w", i, err)
		}

		v, err := strconv.ParseUint(part[130:132], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid v value: %w", i, err)
		}

		result[i] = bindings.Signature{
			V: uint8(v),
			R: rBytes,
			S: sBytes,
		}
	}

	return result, nil
}
