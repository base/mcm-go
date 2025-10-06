package flags

import (
	"github.com/base/mcm-go/pkg/cli"

	ucli "github.com/urfave/cli/v2"
)

// OnchainReadFlags returns the flags needed for on-chain read operations
func OnchainReadFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.StringFlag{
			Name:     "rpc",
			Aliases:  []string{"r"},
			Usage:    "Solana RPC endpoint URL",
			EnvVars:  []string{"MCM_RPC_URL"},
			Required: true,
		},
		&ucli.StringFlag{
			Name:     "program-id",
			Aliases:  []string{"p"},
			Usage:    "MCM program ID (base58)",
			EnvVars:  []string{"MCM_PROGRAM_ID"},
			Required: true,
		},
	}
}

// OnchainWriteFlags returns additional flags needed for on-chain write operations (transactions)
// These flags should be combined with OnchainReadFlags for write commands
func OnchainWriteFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.StringFlag{
			Name:     "ws",
			Usage:    "Solana WebSocket endpoint URL (required for confirmations)",
			EnvVars:  []string{"MCM_WS_URL"},
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
