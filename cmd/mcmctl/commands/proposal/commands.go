package proposal

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the proposal command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "proposal",
		Usage: "Proposal operations (offline and online)",
		Subcommands: []*ucli.Command{
			CreateCommand(),
			HashCommand(),
			SetRootCommand(),
			ExecuteCommand(),
		},
	}
}
