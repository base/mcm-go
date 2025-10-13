package mcm

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the mcm proposal command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "mcm",
		Usage: "MCM configuration proposal operations",
		Subcommands: []*ucli.Command{
			UpdateSignersCommand(),
			AcceptOwnershipCommand(),
		},
	}
}
