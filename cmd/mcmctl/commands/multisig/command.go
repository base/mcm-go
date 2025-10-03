package multisig

import (
	"mcm-go/cmd/mcmctl/flags"

	ucli "github.com/urfave/cli/v2"
)

// Command returns the multisig command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "multisig",
		Usage: "Multisig operations",
		Flags: flags.TransactionFlags(),
		Subcommands: []*ucli.Command{
			InitCommand(),
		},
	}
}
