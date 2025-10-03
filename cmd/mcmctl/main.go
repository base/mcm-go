package main

import (
	"fmt"
	"os"

	"mcm-go/cmd/mcmctl/commands/multisig"
	"mcm-go/cmd/mcmctl/commands/proposal"
	"mcm-go/cmd/mcmctl/commands/signatures"
	"mcm-go/cmd/mcmctl/commands/signers"

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
