package flags

import (
	"mcm-go/pkg/cli"

	ucli "github.com/urfave/cli/v2"
)

// TransactionFlags returns the flags needed for blockchain operations
func TransactionFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.StringFlag{
			Name:     "rpc",
			Aliases:  []string{"r"},
			Usage:    "Solana RPC endpoint URL",
			EnvVars:  []string{"MCM_RPC_URL"},
			Required: true,
		},
		&ucli.StringFlag{
			Name:     "ws",
			Usage:    "Solana WebSocket endpoint URL (required for confirmations)",
			EnvVars:  []string{"MCM_WS_URL"},
			Required: true,
		},
		&ucli.StringFlag{
			Name:     "program-id",
			Aliases:  []string{"p"},
			Usage:    "MCM program ID (base58)",
			EnvVars:  []string{"MCM_PROGRAM_ID"},
			Required: true,
		},
		&ucli.StringFlag{
			Name:    "authority",
			Aliases: []string{"a"},
			Usage:   "Path to authority keypair file (JSON or base58, also used as transaction payer)",
			EnvVars: []string{"MCM_AUTHORITY"},
			Value:   cli.DefaultKeypairPath(),
		},
	}
}
