package loader_v3

import (
	ucli "github.com/urfave/cli/v2"
)

// Command returns the loader v3 proposal command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "loader-v3",
		Usage: "BPF Loader v3 proposal operations",
		Subcommands: []*ucli.Command{
			UpgradeCommand(),
		},
	}
}
