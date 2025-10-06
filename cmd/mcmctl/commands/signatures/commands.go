package signatures

import (
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// Command returns the signatures command group
func Command() *ucli.Command {
	return &ucli.Command{
		Name:  "signatures",
		Usage: "Signature management for setting roots",
		Subcommands: []*ucli.Command{
			InitCommand(),
			AppendCommand(),
			FinalizeCommand(),
			ClearCommand(),
		},
	}
}

// loadSignaturesService loads the signatures service from CLI flags
func loadSignaturesService(c *ucli.Context) (*services.SignaturesService, error) {
	mcmClient, err := util.LoadClient(c)
	if err != nil {
		return nil, err
	}

	return services.NewSignaturesService(mcmClient), nil
}
