package proposal

import (
	"github.com/base/mcm-go/cmd/mcmctl/commands/proposal/bridge"
	"github.com/base/mcm-go/cmd/mcmctl/commands/proposal/loader_v3"
	"github.com/base/mcm-go/cmd/mcmctl/commands/proposal/mcm"

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
			bridge.Command(),
			loader_v3.Command(),
			mcm.Command(),
		},
	}
}
