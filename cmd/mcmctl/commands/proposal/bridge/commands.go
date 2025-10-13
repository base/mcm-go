package bridge

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the bridge proposal command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "bridge",
		Usage: "Bridge proposal operations",
		Subcommands: []*ucli.Command{
			PauseCommand(),
			SetPartnerOracleConfigCommand(),
		},
	}
}
