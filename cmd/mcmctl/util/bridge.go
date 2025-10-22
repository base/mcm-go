package util

import (
	solana "github.com/gagliardetto/solana-go"
)

// BridgePDA derives the PDA for the bridge account
// This is specific to the bridge program and uses the "bridge" seed
func BridgePDA(bridgeProgram solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("bridge"),
		},
		bridgeProgram,
	)
}
