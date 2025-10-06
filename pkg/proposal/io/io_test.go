package io

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hexutil "github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/proposal"

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

	// Verify that the JSON file contains 0x prefixes
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}
	jsonStr := string(jsonData)

	// Check that multisigId has 0x prefix (account for JSON formatting with spaces)
	if !strings.Contains(jsonStr, `"multisigId": "0x`) && !strings.Contains(jsonStr, `"multisigId":"0x`) {
		t.Errorf("JSON multisigId should have 0x prefix. JSON content:\n%s", jsonStr)
	}

	// Check that instruction data has 0x prefix (account for JSON formatting with spaces)
	if !strings.Contains(jsonStr, `"data": "0x`) && !strings.Contains(jsonStr, `"data":"0x`) {
		t.Errorf("JSON instruction data should have 0x prefix. JSON content:\n%s", jsonStr)
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

func TestDecodeHexWith0xPrefix(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantHex string
		wantErr bool
	}{
		{
			name:    "without 0x prefix should fail",
			input:   "deadbeef",
			wantErr: true,
		},
		{
			name:    "with lowercase 0x prefix",
			input:   "0xdeadbeef",
			wantHex: "deadbeef",
			wantErr: false,
		},
		{
			name:    "with uppercase 0X prefix should fail",
			input:   "0XDEADBEEF",
			wantErr: true,
		},
		{
			name:    "32 bytes without prefix should fail",
			input:   "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr: true,
		},
		{
			name:    "32 bytes with prefix",
			input:   "0x0000000000000000000000000000000000000000000000000000000000000000",
			wantHex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr: false,
		},
		{
			name:    "invalid hex",
			input:   "0xzzzz",
			wantHex: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hexutil.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotHex := hex.EncodeToString(got)
				if gotHex != tt.wantHex {
					t.Errorf("decodeHex() = %v, want %v", gotHex, tt.wantHex)
				}
			}
		})
	}
}

func TestLoadInstructions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "instructions.json")

	// Create test instructions file
	content := `[
  {
    "programId": "11111111111111111111111111111111",
    "data": "0xdeadbeef",
    "accounts": [
      {
        "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        "isSigner": true,
        "isWritable": false
      }
    ]
  },
  {
    "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
    "data": "0x1234",
    "accounts": []
  }
]`

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load instructions
	instructions, err := LoadInstructions(filePath)
	if err != nil {
		t.Fatalf("LoadInstructions failed: %v", err)
	}

	if len(instructions) != 2 {
		t.Errorf("Expected 2 instructions, got %d", len(instructions))
	}

	// Check first instruction
	if instructions[0].ProgID.String() != "11111111111111111111111111111111" {
		t.Errorf("Instruction 0 ProgID mismatch")
	}
	if string(instructions[0].DataBytes) != string([]byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Errorf("Instruction 0 Data mismatch")
	}
	if len(instructions[0].AccountValues) != 1 {
		t.Errorf("Instruction 0 accounts count mismatch")
	}
	if !instructions[0].AccountValues[0].IsSigner {
		t.Error("Account should be signer")
	}
	if instructions[0].AccountValues[0].IsWritable {
		t.Error("Account should not be writable")
	}

	// Check second instruction
	if instructions[1].ProgID.String() != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Errorf("Instruction 1 ProgID mismatch")
	}
	if string(instructions[1].DataBytes) != string([]byte{0x12, 0x34}) {
		t.Errorf("Instruction 1 Data mismatch")
	}
	if len(instructions[1].AccountValues) != 0 {
		t.Errorf("Instruction 1 should have no accounts")
	}
}

func TestLoadInstructions_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.json")

	content := `[]`
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadInstructions(filePath)
	if err == nil {
		t.Error("Expected error for empty instructions, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "no instructions") {
		t.Errorf("Expected 'no instructions' error, got: %v", err)
	}
}

func TestLoadInstructions_InvalidHex(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")

	// Missing 0x prefix
	content := `[
  {
    "programId": "11111111111111111111111111111111",
    "data": "deadbeef",
    "accounts": []
  }
]`

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadInstructions(filePath)
	if err == nil {
		t.Error("Expected error for invalid hex, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "0x") {
		t.Errorf("Expected '0x' in error message, got: %v", err)
	}
}
