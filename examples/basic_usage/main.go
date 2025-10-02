package main

import (
	"context"
	"fmt"
	"log"

	"mcm-go/pkg/client"
	"mcm-go/pkg/services"

	"github.com/gagliardetto/solana-go"
)

// This example demonstrates creating proposals from on-chain state,
// computing Merkle roots, and preparing hashes for signing.
func main() {
	// 1. Create MCM client
	cfg := client.Config{
		RPCURL:    "https://api.devnet.solana.com",
		WSURL:     "wss://api.devnet.solana.com",
		ProgramID: solana.MustPublicKeyFromBase58("55CNTEUq6cAa2sBA7bkDfJ2bb3uWs7Zh77vAF9H8TnJL"),
	}

	mcmClient, err := client.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer mcmClient.Close()

	// 2. Create proposal service
	proposalSvc := services.NewProposalService(mcmClient)

	// 3. Define multisig parameters
	var multisigID [32]byte
	copy(multisigID[:], []byte("my-multisig-001"))

	instructions := []*solana.GenericInstruction{
		{
			ProgID:        solana.SystemProgramID,
			DataBytes:     []byte{0x02, 0x00, 0x00, 0x00},
			AccountValues: []*solana.AccountMeta{},
		},
	}

	validUntil := uint32(1800000000)

	// 4. Create proposal from on-chain state
	fmt.Println("📡 Creating proposal from on-chain state...")
	ctx := context.Background()
	p, err := proposalSvc.CreateProposalFromChain(ctx, services.CreateProposalFromChainParams{
		MultisigID:           multisigID,
		ValidUntil:           validUntil,
		Instructions:         instructions,
		OverridePreviousRoot: false,
	})
	if err != nil {
		log.Fatalf("Failed to create proposal from chain: %v\nNote: This requires an existing on-chain multisig", err)
	}

	fmt.Printf("✅ Proposal created successfully!\n")
	fmt.Printf("   Valid Until: %d\n", p.ValidUntil)
	fmt.Printf("   Instructions: %d\n", len(p.Instructions))
	fmt.Printf("   Chain ID: %d\n", p.RootMetadata.ChainID)
	fmt.Printf("   Pre Op Count: %d\n", p.RootMetadata.PreOpCount)
	fmt.Printf("   Post Op Count: %d\n", p.RootMetadata.PostOpCount)

	// 5. Compute Merkle root and proofs
	fmt.Println("\n📊 Computing Merkle root...")
	proposalWithRoot, err := p.WithRoot()
	if err != nil {
		log.Fatalf("Failed to compute root: %v", err)
	}
	fmt.Printf("✅ Merkle Root: 0x%x\n", proposalWithRoot.Root)

	// 6. Compute hash to sign
	fmt.Println("\n🔏 Computing hash to sign...")
	proposalToSign, err := proposalWithRoot.WithHashToSign()
	if err != nil {
		log.Fatalf("Failed to compute hash to sign: %v", err)
	}

	fmt.Printf("✅ Hash to sign: 0x%x\n", proposalToSign.HashToSign)
	fmt.Printf("   This hash should be distributed to signers for ECDSA signing\n")

	// 7. Next steps
	fmt.Println("\n📝 Next steps in a real scenario:")
	fmt.Println("   1. Distribute proposalToSign.HashToSign to all signers")
	fmt.Println("   2. Collect ECDSA signatures using SignaturesService")
	fmt.Println("   3. Submit the root on-chain with ProposalService.SetRoot()")
	fmt.Println("   4. Execute operations using ExecutionService")
	fmt.Println("   5. Save/load proposals using pkg/proposal/io (see persistence example)")

	fmt.Println("\n✅ Example completed successfully!")
}
