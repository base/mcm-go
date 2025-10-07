package signers

import (
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// Command returns the signers command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "signers",
		Usage: "Signers management",
		Subcommands: []*ucli.Command{
			InitCommand(),
			AppendCommand(),
			FinalizeCommand(),
			ClearCommand(),
			SetConfigCommand(),
		},
	}
}

// loadSignersService loads the signers service from CLI flags
func loadSignersService(c *ucli.Context) (*services.SignersService, error) {
	mcmClient, err := util.LoadClient(c)
	if err != nil {
		return nil, err
	}

	return services.NewSignersService(mcmClient), nil
}
