package signers

import (
	"fmt"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/pkg/cli"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// SetConfigCommand returns the signers set-config command
func SetConfigCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "set-config",
		Usage: "Set the multisig configuration (signer groups and quorums)",
		Flags: append(flags.OnchainWriteFlags(),
			flags.MultisigIDFlag(),
			flags.SignerGroupsFlag(),
			flags.GroupQuorumsFlag(),
			flags.GroupParentsFlag(),
			flags.ClearRootFlag(),
		),
		Action: func(c *ucli.Context) error {
			svc, err := loadSignersService(c)
			if err != nil {
				return err
			}

			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Parse signer groups
			signerGroups, err := cli.ParseUint8Slice(c.String("signer-groups"))
			if err != nil {
				return fmt.Errorf("invalid signer-groups: %w", err)
			}

			// Parse and pad group quorums to 32 elements
			groupQuorums, err := cli.ParseAndPadUint8Array32(c.String("group-quorums"))
			if err != nil {
				return fmt.Errorf("invalid group-quorums: %w", err)
			}

			// Parse and pad group parents to 32 elements
			groupParents, err := cli.ParseAndPadUint8Array32(c.String("group-parents"))
			if err != nil {
				return fmt.Errorf("invalid group-parents: %w", err)
			}

			clearRoot := c.Bool("clear-root")

			sig, err := svc.SetConfig(c.Context, services.SetConfigParams{
				MultisigID:   multisigID,
				SignerGroups: signerGroups,
				GroupQuorums: groupQuorums,
				GroupParents: groupParents,
				ClearRoot:    clearRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}

			fmt.Printf("configuration set successfully\n")
			fmt.Printf("  signer groups: %v\n", signerGroups)
			fmt.Printf("  group quorums: %v (first %d)\n", groupQuorums[:min(5, len(groupQuorums))], min(5, len(groupQuorums)))
			fmt.Printf("  group parents: %v (first %d)\n", groupParents[:min(5, len(groupParents))], min(5, len(groupParents)))
			fmt.Printf("  clear root: %v\n", clearRoot)
			fmt.Printf("signature: %s\n", sig)
			return nil
		},
	}
}
