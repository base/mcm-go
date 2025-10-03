package signers

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// loadSignersService loads the signers service from CLI flags
func loadSignersService(c *ucli.Context) (*services.SignersService, error) {
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc"),
		WSUrl:       c.String("ws"),
		ProgramID:   c.String("program-id"),
		KeypairPath: c.String("authority"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	mcmClient, err := client.New(*cfg)
	if err != nil {
		return nil, err
	}

	return services.NewSignersService(mcmClient), nil
}
