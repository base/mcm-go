package signers

import (
	"mcm-go/cmd/mcmctl/flags"

	ucli "github.com/urfave/cli/v2"
)

// Command returns the signers command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "signers",
		Usage: "Signers management",
		Flags: flags.TransactionFlags(),
		Subcommands: []*ucli.Command{
			InitCommand(),
			AppendCommand(),
			FinalizeCommand(),
			SetConfigCommand(),
		},
	}
}
