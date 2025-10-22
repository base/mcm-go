package util

import (
	"fmt"

	"github.com/base/mcm-go/pkg/cli"
	"github.com/base/mcm-go/pkg/client"

	ucli "github.com/urfave/cli/v2"
)

// LoadClient loads the MCM client from CLI flags
// WS and authority are optional (only needed for write operations)
func LoadClient(c *ucli.Context) (*client.Client, error) {
	cfg, err := cli.LoadConfig(cli.ConfigParams{
		RPCUrl:      c.String("rpc-url"),
		WSUrl:       c.String("ws-url"),
		ProgramID:   c.String("mcm-program-id"),
		KeypairPath: c.String("authority"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return client.New(*cfg)
}
