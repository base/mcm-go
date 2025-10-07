package multisig

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the multisig command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "multisig",
		Usage: "Multisig operations",
		Subcommands: []*ucli.Command{
			InitCommand(),
			PrintConfigCommand(),
			PrintAuthorityCommand(),
		},
	}
}
