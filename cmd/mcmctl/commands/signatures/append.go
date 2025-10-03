package signatures

import (
	"fmt"
	"strconv"
	"strings"

	"mcm-go/pkg/bindings"
	"mcm-go/pkg/cli"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// AppendCommand returns the signatures append command
func AppendCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "append",
		Usage: "Append ECDSA signatures to storage",
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:     "multisig-id",
				Usage:    "Multisig identifier (32 bytes hex)",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "root",
				Usage:    "Merkle root (32 bytes hex)",
				Required: true,
			},
			&ucli.Uint64Flag{
				Name:     "valid-until",
				Usage:    "Unix timestamp until which the root is valid",
				Required: true,
			},
			&ucli.StringFlag{
				Name:     "signatures",
				Usage:    "Comma-separated ECDSA signatures in format 'v:r:s' (hex values, with/without 0x)",
				Required: true,
			},
		},
		Action: func(c *ucli.Context) error {
			svc, err := loadSignaturesService(c)
			if err != nil {
				return err
			}

			multisigID, err := cli.ParseHex32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			root, err := cli.ParseHex32(c.String("root"))
			if err != nil {
				return fmt.Errorf("invalid root: %w", err)
			}

			validUntil, err := parseValidUntil(c.Uint64("valid-until"))
			if err != nil {
				return err
			}

			// Parse signatures
			sigsBatch, err := parseSignatures(c.String("signatures"))
			if err != nil {
				return fmt.Errorf("invalid signatures: %w", err)
			}

			sig, err := svc.AppendSignatures(c.Context, services.AppendSignaturesParams{
				MultisigID:      multisigID,
				Root:            root,
				ValidUntil:      validUntil,
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

// parseSignatures parses comma-separated ECDSA signatures in format "v:r:s"
func parseSignatures(s string) ([]bindings.Signature, error) {
	if s == "" {
		return nil, fmt.Errorf("empty signatures input")
	}

	parts := strings.Split(s, ",")
	result := make([]bindings.Signature, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		components := strings.Split(part, ":")
		if len(components) != 3 {
			return nil, fmt.Errorf("signature %d: expected format 'v:r:s', got %q", i, part)
		}

		// Parse v (recovery id, 0-3)
		vStr := strings.TrimSpace(components[0])
		vStr = strings.TrimPrefix(vStr, "0x")
		vStr = strings.TrimPrefix(vStr, "0X")
		v, err := strconv.ParseUint(vStr, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid v value %q: %w", i, components[0], err)
		}

		// Parse r (32 bytes)
		rBytes, err := cli.ParseHex32(components[1])
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid r value: %w", i, err)
		}

		// Parse s (32 bytes)
		sBytes, err := cli.ParseHex32(components[2])
		if err != nil {
			return nil, fmt.Errorf("signature %d: invalid s value: %w", i, err)
		}

		result[i] = bindings.Signature{
			V: uint8(v),
			R: rBytes,
			S: sBytes,
		}
	}

	return result, nil
}
