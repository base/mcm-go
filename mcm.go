// Package mcm provides a comprehensive Go SDK for interacting with the Multi-Chain Multisig (MCM) program on Solana.
//
// The SDK is organized into several packages:
//   - client: RPC client wrapper with MCM-specific configuration
//   - pda: Program Derived Address (PDA) derivation utilities
//   - crypto: Cryptographic utilities (Merkle trees, ECDSA signatures)
//   - proposal: Proposal creation and management
//   - instructions: High-level instruction builders
//   - state: On-chain account fetchers
//   - services: High-level services for common workflows
//   - tx: Transaction building and submission utilities
//
// Example usage:
//
//	import (
//	    "context"
//	    "github.com/gagliardetto/solana-go"
//	    "mcm-go/pkg/client"
//	    "mcm-go/pkg/services"
//	)
//
//	func main() {
//	    // Create client
//	    cfg := client.Config{
//	        RPCURL:    "https://api.devnet.solana.com",
//	        ProgramID: solana.MustPublicKeyFromBase58("55CNTEUq6cAa2sBA7bkDfJ2bb3uWs7Zh77vAF9H8TnJL"),
//	    }
//	    mcmClient, _ := client.New(cfg)
//
//	    // Create services
//	    proposalSvc := services.NewProposalService(mcmClient)
//
//	    // Create and submit proposals
//	    // ...
//	}
package mcm

import (
	"mcm-go/pkg/bindings"
	"mcm-go/pkg/client"
	"mcm-go/pkg/crypto"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/pda"
	"mcm-go/pkg/proposal"
	"mcm-go/pkg/services"
	"mcm-go/pkg/state"
	"mcm-go/pkg/tx"
)

// Re-export commonly used types and functions for convenience

// Client types
type (
	Client       = client.Client
	ClientConfig = client.Config
)

// Proposal types
type (
	Proposal         = proposal.Proposal
	ProposalWithRoot = proposal.ProposalWithRoot
	ProposalToSign   = proposal.ProposalToSign
	ProposalBuilder  = proposal.Builder
	RootMetadata     = proposal.RootMetadata
)

// Service types
type (
	SignersService    = services.SignersService
	SignaturesService = services.SignaturesService
	ProposalService   = services.ProposalService
	ExecutionService  = services.ExecutionService
)

// Service parameter types
type (
	InitSignersParams               = services.InitSignersParams
	AppendSignersParams             = services.AppendSignersParams
	AppendSignersInBatchesParams    = services.AppendSignersInBatchesParams
	FinalizeSignersParams           = services.FinalizeSignersParams
	InitSignaturesParams            = services.InitSignaturesParams
	AppendSignaturesParams          = services.AppendSignaturesParams
	AppendSignaturesInBatchesParams = services.AppendSignaturesInBatchesParams
	FinalizeSignaturesParams        = services.FinalizeSignaturesParams
	SetRootParams                   = services.SetRootParams
	CreateProposalFromChainParams   = services.CreateProposalFromChainParams
	ExecuteOperationParams          = services.ExecuteOperationParams
	ExecuteAllOperationsParams      = services.ExecuteAllOperationsParams
)

// State types
type (
	StateFetcher = state.Fetcher
)

// Crypto types
type (
	MerkleProof          = crypto.Proof
	MerkleRootWithProofs = crypto.MerkleRootWithProofs
	ECDSASignature       = crypto.ECDSASignature
)

// Transaction types
type (
	TransactionBuilder = tx.Builder
)

// Bindings types (re-exported from generated code)
type (
	MultisigConfig         = bindings.MultisigConfig
	ConfigSigners          = bindings.ConfigSigners
	RootSignatures         = bindings.RootSignatures
	RootMetadataAccount    = bindings.RootMetadata
	ExpiringRootAndOpCount = bindings.ExpiringRootAndOpCount
	SeenSignedHash         = bindings.SeenSignedHash
	McmSigner              = bindings.McmSigner
	Signature              = bindings.Signature
)

// Commonly used functions

// NewClient creates a new MCM client
var NewClient = client.New

// NewProposalBuilder creates a new proposal builder
var NewProposalBuilder = proposal.NewBuilder

// BuildMerkleTree builds a Merkle tree from leaves
var BuildMerkleTree = crypto.BuildMerkleTreeFromLeaves

// PDA derivation functions
var (
	MultisigSignerPDA         = pda.MultisigSignerPDA
	MultisigConfigPDA         = pda.MultisigConfigPDA
	MultisigConfigSignersPDA  = pda.MultisigConfigSignersPDA
	RootMetadataPDA           = pda.RootMetadataPDA
	ExpiringRootAndOpCountPDA = pda.ExpiringRootAndOpCountPDA
	RootSignaturesPDA         = pda.RootSignaturesPDA
	SeenSignedHashesPDA       = pda.SeenSignedHashesPDA
	ProgramDataPDA            = pda.ProgramDataPDA
)

// Instruction builders
var (
	Initialize         = instructions.Initialize
	SetConfig          = instructions.SetConfig
	SetRoot            = instructions.SetRoot
	Execute            = instructions.Execute
	InitSigners        = instructions.InitSigners
	AppendSigners      = instructions.AppendSigners
	FinalizeSigners    = instructions.FinalizeSigners
	InitSignatures     = instructions.InitSignatures
	AppendSignatures   = instructions.AppendSignatures
	FinalizeSignatures = instructions.FinalizeSignatures
)

// Service constructors
var (
	NewSignersService    = services.NewSignersService
	NewSignaturesService = services.NewSignaturesService
	NewProposalService   = services.NewProposalService
	NewExecutionService  = services.NewExecutionService
)

// State fetchers
var NewStateFetcher = state.NewFetcher

// Transaction builder
var NewTxBuilder = tx.New

// Version information
const (
	Version = "0.1.0"
)
