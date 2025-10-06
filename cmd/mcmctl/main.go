package main

import (
	"fmt"
	"os"

	"github.com/base/mcm-go/cmd/mcmctl/commands/multisig"
	"github.com/base/mcm-go/cmd/mcmctl/commands/proposal"
	"github.com/base/mcm-go/cmd/mcmctl/commands/signatures"
	"github.com/base/mcm-go/cmd/mcmctl/commands/signers"

	ucli "github.com/urfave/cli/v2"
)

func main() {
	app := &ucli.App{
		Name:  "mcmctl",
		Usage: "CLI tool for managing MCM multisig on Solana",
		Commands: []*ucli.Command{
			multisig.Command(),
			signers.Command(),
			proposal.Command(),
			signatures.Command(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
