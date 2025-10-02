package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mcm-go/pkg/proposal"
	"mcm-go/pkg/proposal/io"

	"github.com/gagliardetto/solana-go"
)

// This example demonstrates saving and loading proposals to/from JSON files.
func main() {
	// 1. Create a proposal
	var multisigID [32]byte
	copy(multisigID[:], []byte("example-multisig"))

	p := &proposal.Proposal{
		MultisigID: multisigID,
		ValidUntil: 1800000000,
		Instructions: []*solana.GenericInstruction{
			{
				ProgID:    solana.SystemProgramID,
				DataBytes: []byte{0x02, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
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
			Multisig:             solana.MustPublicKeyFromBase58("11111111111111111111111111111112"),
			PreOpCount:           0,
			PostOpCount:          1,
			OverridePreviousRoot: false,
		},
	}

	// 2. Save to JSON
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "proposal.json")

	fmt.Printf("💾 Saving proposal to: %s\n", filePath)
	if err := io.SaveProposal(p, filePath); err != nil {
		log.Fatalf("Failed to save proposal: %v", err)
	}
	fmt.Println("✅ Proposal saved successfully!")

	// 3. Load from JSON
	fmt.Printf("\n📂 Loading proposal from: %s\n", filePath)
	loaded, err := io.LoadProposal(filePath)
	if err != nil {
		log.Fatalf("Failed to load proposal: %v", err)
	}
	fmt.Println("✅ Proposal loaded successfully!")

	// 4. Verify loaded data
	fmt.Printf("\n📋 Loaded Proposal Details:\n")
	fmt.Printf("   MultisigID: %x\n", loaded.MultisigID)
	fmt.Printf("   Valid Until: %d\n", loaded.ValidUntil)
	fmt.Printf("   Instructions: %d\n", len(loaded.Instructions))
	fmt.Printf("   Chain ID: %d\n", loaded.RootMetadata.ChainID)
	fmt.Printf("   Pre Op Count: %d\n", loaded.RootMetadata.PreOpCount)
	fmt.Printf("   Post Op Count: %d\n", loaded.RootMetadata.PostOpCount)

	// 5. Compute Merkle root and hash to sign
	fmt.Printf("\n🔏 Computing Merkle root and hash to sign:\n")
	proposalWithRoot, err := loaded.WithRoot()
	if err != nil {
		log.Fatalf("Failed to compute root: %v", err)
	}
	fmt.Printf("   Root: 0x%x\n", proposalWithRoot.Root)

	proposalToSign, err := proposalWithRoot.WithHashToSign()
	if err != nil {
		log.Fatalf("Failed to compute hash to sign: %v", err)
	}
	fmt.Printf("   Hash to sign: 0x%x\n", proposalToSign.HashToSign)
	fmt.Printf("   Metadata Proof: %d hashes\n", len(proposalToSign.MetadataProof))
	fmt.Printf("   Operation Proofs: %d proofs\n", len(proposalToSign.OperationProofs))

	// 6. Display JSON file content
	fmt.Printf("\n📄 JSON File Content:\n")
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	fmt.Println(string(content))

	// 7. Cleanup
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: failed to cleanup temp file: %v", err)
	}

	fmt.Println("\n✅ Persistence example completed successfully!")
}
