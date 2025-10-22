// Package pda provides utilities for deriving Program Derived Addresses (PDAs)
// for the Bridge program on Solana.
package pda

import (
	"github.com/gagliardetto/solana-go"
)

var (
	// Seed constants
	bridgeSeed = []byte("bridge")
)

// BridgePDA derives the PDA for the bridge account
func BridgePDA(bridgeProgram solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			bridgeSeed,
		},
		bridgeProgram,
	)
}
