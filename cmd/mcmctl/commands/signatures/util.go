package signatures

import (
	"fmt"

	"mcm-go/pkg/cli"
	"mcm-go/pkg/client"
	"mcm-go/pkg/services"

	ucli "github.com/urfave/cli/v2"
)

// loadSignaturesService loads the signatures service from CLI flags
func loadSignaturesService(c *ucli.Context) (*services.SignaturesService, error) {
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

	return services.NewSignaturesService(mcmClient), nil
}

// parseValidUntil validates that a uint64 fits in uint32
func parseValidUntil(validUntil uint64) (uint32, error) {
	if validUntil > 4294967295 {
		return 0, fmt.Errorf("valid-until must fit in uint32")
	}
	return uint32(validUntil), nil
}
