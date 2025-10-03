package instructions

import (
	"testing"

	"github.com/gagliardetto/solana-go"
)

// TestSetProgramID verifies that setProgramID correctly overrides the ProgramID
func TestSetProgramID(t *testing.T) {
	// Create a test instruction
	accounts := solana.AccountMetaSlice{}
	accounts.Append(solana.NewAccountMeta(solana.MustPublicKeyFromBase58("11111111111111111111111111111111"), false, false))

	originalProgramID := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	newProgramID := solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")

	ix := solana.NewInstruction(originalProgramID, accounts, []byte{})

	// Verify original ProgramID
	if ix.ProgramID() != originalProgramID {
		t.Errorf("Expected original ProgramID %s, got %s", originalProgramID, ix.ProgramID())
	}

	// Apply setProgramID
	modifiedIx := setProgramID(ix, newProgramID)

	// Verify ProgramID was overridden
	if modifiedIx.ProgramID() != newProgramID {
		t.Errorf("Expected new ProgramID %s, got %s", newProgramID, modifiedIx.ProgramID())
	}
}
