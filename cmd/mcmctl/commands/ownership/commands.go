package ownership

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the ownership command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "ownership",
		Usage: "Ownership management operations",
		Subcommands: []*ucli.Command{
			TransferOwnershipCommand(),
			AcceptOwnershipCommand(),
		},
	}
}
