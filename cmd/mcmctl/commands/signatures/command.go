package signatures

import (
	"mcm-go/cmd/mcmctl/flags"

	ucli "github.com/urfave/cli/v2"
)

// Command returns the signatures command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "signatures",
		Usage: "Signature management for setting roots",
		Flags: flags.TransactionFlags(),
		Subcommands: []*ucli.Command{
			InitCommand(),
			AppendCommand(),
			FinalizeCommand(),
			ClearCommand(),
		},
	}
}
