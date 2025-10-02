package io

import (
	"os"
	"path/filepath"
	"testing"

	"mcm-go/pkg/proposal"

	"github.com/gagliardetto/solana-go"
)

func TestSaveLoadProposal(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "proposal.json")

	// Create a test proposal
	var multisigID [32]byte
	copy(multisigID[:], []byte("test-multisig-001"))

	original := &proposal.Proposal{
		MultisigID: multisigID,
		ValidUntil: 1800000000,
		Instructions: []*solana.GenericInstruction{
			{
				ProgID:    solana.SystemProgramID,
				DataBytes: []byte{0xde, 0xad, 0xbe, 0xef},
				AccountValues: []*solana.AccountMeta{
					{
						PublicKey:  solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
						IsSigner:   true,
						IsWritable: false,
					},
				},
			},
		},
		RootMetadata: proposal.RootMetadata{
			ChainID:              1,
			Multisig:             solana.MustPublicKeyFromBase58("11111111111111111111111111111111"),
			PreOpCount:           0,
			PostOpCount:          1,
			OverridePreviousRoot: false,
		},
	}

	// Save
	if err := SaveProposal(original, filePath); err != nil {
		t.Fatalf("SaveProposal failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Load
	loaded, err := LoadProposal(filePath)
	if err != nil {
		t.Fatalf("LoadProposal failed: %v", err)
	}

	// Compare
	if loaded.MultisigID != original.MultisigID {
		t.Errorf("MultisigID mismatch: got %v, want %v", loaded.MultisigID, original.MultisigID)
	}
	if loaded.ValidUntil != original.ValidUntil {
		t.Errorf("ValidUntil mismatch: got %d, want %d", loaded.ValidUntil, original.ValidUntil)
	}
	if len(loaded.Instructions) != len(original.Instructions) {
		t.Fatalf("Instructions count mismatch: got %d, want %d", len(loaded.Instructions), len(original.Instructions))
	}

	// Check instruction
	if !loaded.Instructions[0].ProgID.Equals(original.Instructions[0].ProgID) {
		t.Errorf("Instruction ProgID mismatch")
	}
	if string(loaded.Instructions[0].DataBytes) != string(original.Instructions[0].DataBytes) {
		t.Errorf("Instruction Data mismatch")
	}

	// Check metadata
	if loaded.RootMetadata.ChainID != original.RootMetadata.ChainID {
		t.Errorf("ChainID mismatch")
	}
	if !loaded.RootMetadata.Multisig.Equals(original.RootMetadata.Multisig) {
		t.Errorf("Multisig mismatch")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(filePath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadProposal(filePath)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := LoadProposal("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}
