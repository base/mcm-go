# MCM Go SDK

Go SDK for the Multi-Chain Multisig (MCM) Solana program.

## CLI Tool

The SDK includes `mcmctl`, a command-line tool for managing MCM multisigs:

```bash
go install github.com/base/mcm-go/cmd/mcmctl@latest
```

```bash
# Set environment variables
export RPC_URL="devnet"
export WS_URL="devnet"
export MCM_PROGRAM_ID="YourProgramID"

# Initialize a multisig (hex values must use 0x prefix)
mcmctl multisig init --multisig-id <hex32> --chain-id 1

# Manage signers
mcmctl signers init --multisig-id <hex32> --total 10
mcmctl signers append --multisig-id <hex32> --signers <addr1>,<addr2>,...
mcmctl signers finalize --multisig-id <hex32>
mcmctl signers clear --multisig-id <hex32>
mcmctl signers set-config --multisig-id <hex32> --signer-groups <groups> --group-quorums <quorums> --group-parents <parents>
```

See [cmd/mcmctl/README.md](cmd/mcmctl/README.md) for complete documentation.

## Quick Start

```go
package main

import (
    "context"
    "github.com/gagliardetto/solana-go"
    "github.com/base/mcm-go/pkg/client"
    "github.com/base/mcm-go/pkg/proposal"
    "github.com/base/mcm-go/pkg/services"
)

func main() {
    // Setup client
    payer := solana.MustPrivateKeyFromBase58("your-private-key")
    programID := solana.MustPublicKeyFromBase58("YourProgramID")

    cfg := client.Config{
        RPCURL:    "https://api.devnet.solana.com",
        WSURL:     "wss://api.devnet.solana.com",
        ProgramID: programID,
        Payer:     &payer,
    }

    mcmClient, _ := client.New(cfg)
    defer mcmClient.Close()

    // Create proposal from on-chain state
    var multisigID [32]byte // Your multisig ID
    var validUntil uint32 = 1800000000
    var instructions []solana.Instruction // Your instructions

    proposalSvc := services.NewProposalService(mcmClient)
    ctx := context.Background()

    p, _ := proposalSvc.CreateProposalFromChain(ctx, services.CreateProposalFromChainParams{
        MultisigID:           multisigID,
        ValidUntil:           validUntil,
        Instructions:         instructions,
        OverridePreviousRoot: false,
    })

    // Compute Merkle root and hash to sign
    pwr, _ := p.WithRoot()
    pts, _ := pwr.WithHashToSign()

    // Distribute pts.HashToSign to signers for ECDSA signing
}
```

## CLI Examples

The `cmd/mcmctl` directory provides a complete command-line interface demonstrating SDK usage:

- **Multisig operations** - Initialize multisig accounts on Solana
- **Signers management** - Configure signer addresses and groups
- **Signatures management** - Submit ECDSA signatures for proposal approval
- **Proposal operations** - Compute hash for signing, set roots, and execute operations on-chain

See [cmd/mcmctl/README.md](cmd/mcmctl/README.md) for detailed usage examples.

## Package Structure

```
mcm-go/
├── pkg/
│   ├── bindings/      # Anchor-generated types from mcm.json IDL
│   ├── client/        # Solana RPC/WebSocket client wrapper with transaction helpers
│   ├── crypto/        # Keccak256 Merkle tree implementation with proof generation
│   ├── pda/           # Program Derived Address utilities
│   ├── proposal/      # Proposal types, builder, Merkle computation, signing
│   │   ├── io/        # JSON persistence (save/load proposals)
│   │   ├── types.go   # Core types (Proposal, ProposalWithRoot, ProposalToSign)
│   │   ├── builder.go # Builder pattern for constructing proposals
│   │   ├── merkle.go  # Merkle root computation (p.WithRoot())
│   │   └── signing.go # Hash to sign computation (pwr.WithHashToSign())
│   ├── instructions/  # MCM instruction builders (Initialize, SetConfig, etc.)
│   ├── state/         # On-chain account fetchers
│   └── services/      # High-level services (ProposalService, SignersService, etc.)
├── cmd/mcmctl/        # CLI demonstrating SDK usage
└── mcm.json          # MCM program IDL (Anchor >= 0.30.0)
```

## Core Concepts

### 1. Proposals

Proposals contain instructions and metadata. The SDK provides a fluent API for computing cryptographic components:

```go
import "github.com/base/mcm-go/pkg/proposal"

// Option 1: Using Builder
builder := proposal.NewBuilder(multisigID, validUntil)
builder.SetRootMetadata(metadata)
builder.AddInstruction(instruction)
p, _ := builder.Build()

// Option 2: Direct construction
p := &proposal.Proposal{
    MultisigID:   multisigID,
    ValidUntil:   validUntil,
    Instructions: instructions,
    RootMetadata: metadata,
}

// Compute Merkle root and proofs
pwr, _ := p.WithRoot()

// Compute hash for ECDSA signing (keccak256(root || validUntil))
pts, _ := pwr.WithHashToSign()

// Distribute pts.HashToSign to signers
```

**IMPORTANT - Execution Authority:**

When creating proposals, ensure that the account used as `authority` when executing operations is NOT present in the accounts of any proposal operation. If the same account appears in both places, the program may fail with `ProofCannotBeVerified` error due to signer flag inconsistencies during Merkle proof verification.

**Recommended:** Use a dedicated account as execution authority that never appears in operation accounts.

### 2. Merkle Trees

Keccak256-based Merkle tree with automatic proof generation:

```go
import "github.com/base/mcm-go/pkg/crypto"

leaves := [][32]byte{leaf1, leaf2, leaf3}
tree, _ := crypto.BuildMerkleTreeFromLeaves(leaves)

// tree.Root is the Merkle root
// tree.Proofs[i] is the proof for leaves[i]
```

### 3. PDA Derivation

Derive Program Derived Addresses:

```go
import "github.com/base/mcm-go/pkg/pda"

configPDA, _, _ := pda.MultisigConfigPDA(programID, multisigID)
rootMetadataPDA, _, _ := pda.RootMetadataPDA(programID, multisigID)
```

### 4. Services

High-level services for common workflows:

```go
import "github.com/base/mcm-go/pkg/services"

// Signers management
signersSvc := services.NewSignersService(client)
signersSvc.InitSigners(ctx, params)
signersSvc.AppendSigners(ctx, params)
signersSvc.FinalizeSigners(ctx, params)
signersSvc.SetConfig(ctx, params)

// Signatures management
sigsSvc := services.NewSignaturesService(client)
sigsSvc.InitSignatures(ctx, params)
sigsSvc.AppendSignatures(ctx, params)
sigsSvc.FinalizeSignatures(ctx, params)

// Proposal service (includes creation, root setting, and execution)
proposalSvc := services.NewProposalService(client)
p, _ := proposalSvc.CreateProposalFromChain(ctx, params)
proposalSvc.SetRoot(ctx, params)
proposalSvc.Execute(ctx, params) // Execute operations (single, multiple, or all)
```

### 5. Persistence

Save and load proposals to/from JSON:

```go
import "github.com/base/mcm-go/pkg/proposal/io"

// Save proposal to file
io.SaveProposal(p, "proposal.json")

// Load proposal from file
p, _ := io.LoadProposal("proposal.json")

// Compute root and hash after loading
pwr, _ := p.WithRoot()
pts, _ := pwr.WithHashToSign()
```

## Complete Workflow

### 1. Initialize Multisig

```go
import "github.com/base/mcm-go/pkg/instructions"

ix, _ := instructions.Initialize(instructions.InitializeParams{
    ChainID:    1,
    MultisigID: multisigID,
    Authority:  authority,
    ProgramID:  programID,
})
```

### 2. Configure Signers

```go
signersSvc := services.NewSignersService(client)

// Initialize signer storage
signersSvc.InitSigners(ctx, services.InitSignersParams{
    MultisigID:   multisigID,
    TotalSigners: 10,
})

// Add signers
signersSvc.AppendSigners(ctx, services.AppendSignersParams{
    MultisigID:   multisigID,
    SignersBatch: signerAddresses,
})

// Finalize
signersSvc.FinalizeSigners(ctx, services.FinalizeSignersParams{
    MultisigID: multisigID,
})
```

### 3. Set Configuration

```go
ix, _ := instructions.SetConfig(instructions.SetConfigParams{
    MultisigID:   multisigID,
    SignerGroups: groups,
    GroupQuorums: quorums,
    GroupParents: parents,
    ClearRoot:    false,
    Authority:    authority,
    ProgramID:    programID,
})
```

### 4. Create Proposal and Collect Signatures

```go
proposalSvc := services.NewProposalService(client)

// Create proposal from on-chain state
p, _ := proposalSvc.CreateProposalFromChain(ctx, services.CreateProposalFromChainParams{
    MultisigID:           multisigID,
    ValidUntil:           validUntil,
    Instructions:         instructions,
    OverridePreviousRoot: false,
})

// Compute root and hash
pwr, _ := p.WithRoot()
pts, _ := pwr.WithHashToSign()

// Distribute pts.HashToSign to signers for off-chain ECDSA signing
// Collect signatures...

// Submit signatures on-chain
sigsSvc := services.NewSignaturesService(client)
sigsSvc.InitSignatures(ctx, services.InitSignaturesParams{
    MultisigID:      multisigID,
    Root:            pwr.Root,
    ValidUntil:      validUntil,
    TotalSignatures: uint8(len(signatures)),
})
sigsSvc.AppendSignatures(ctx, services.AppendSignaturesParams{
    MultisigID:      multisigID,
    Root:            pwr.Root,
    ValidUntil:      validUntil,
    SignaturesBatch: signatures,
})
sigsSvc.FinalizeSignatures(ctx, services.FinalizeSignaturesParams{
    MultisigID: multisigID,
    Root:       pwr.Root,
    ValidUntil: validUntil,
})
```

### 5. Set Root and Execute

```go
// Set root on-chain
proposalSvc.SetRoot(ctx, services.SetRootParams{
    MultisigID: multisigID,
    Proposal:   pwr,
})

// Execute all operations
proposalSvc.Execute(ctx, services.ExecuteParams{
    MultisigID:       multisigID,
    ProposalWithRoot: pwr,
    StartIndex:       0,
    OperationCount:   len(pwr.Instructions),
})

// Execute first operation
proposalSvc.Execute(ctx, services.ExecuteParams{
    MultisigID:       multisigID,
    ProposalWithRoot: pwr,
    StartIndex:       0,
    OperationCount:   1,
})
// Execute next operation
proposalSvc.Execute(ctx, services.ExecuteParams{
    MultisigID:       multisigID,
    ProposalWithRoot: pwr,
    StartIndex:       1,
    OperationCount:   1,
})
// Execute first two operations together
proposalSvc.Execute(ctx, services.ExecuteParams{
    MultisigID:       multisigID,
    ProposalWithRoot: pwr,
    StartIndex:       0,
    OperationCount:   2,
})
```

## State Fetching

Fetch on-chain account state:

```go
import "github.com/base/mcm-go/pkg/state"

fetcher := state.NewFetcher(rpcClient, programID)

config, _ := fetcher.GetMultisigConfig(ctx, multisigID)
rootAndOpCount, _ := fetcher.GetExpiringRootAndOpCount(ctx, multisigID)
rootMetadata, _ := fetcher.GetRootMetadata(ctx, multisigID)
```

## Architecture

The SDK is organized in layers:

1. **Bindings** (`pkg/bindings`) - Anchor-generated types from IDL
2. **Core Utilities** (`pkg/pda`, `pkg/crypto`) - PDAs and Merkle trees
3. **Proposal Layer** (`pkg/proposal`) - Proposal construction and cryptography
4. **Instructions** (`pkg/instructions`) - MCM instruction builders
5. **State** (`pkg/state`) - On-chain account fetchers
6. **Services** (`pkg/services`) - High-level workflows
7. **Client** (`pkg/client`) - RPC, WebSocket, and transaction handling

## Testing

```bash
go test ./...
```

## Dependencies

- [solana-go](https://github.com/gagliardetto/solana-go) - Solana Go SDK
- [anchor-go](https://github.com/gagliardetto/anchor-go) - Used to generate bindings from `mcm.json` IDL

## IDL Source

The `mcm.json` IDL is sourced from the [MCM Solana program](https://github.com/smartcontractkit/chainlink-ccip/blob/main/chains/solana/contracts/target/idl/mcm.json) and updated to align with Anchor >= 0.30.0.

## Links

- [MCM Solana Program](https://github.com/smartcontractkit/chainlink-ccip/tree/main/chains/solana/contracts/programs/mcm)

## License

MIT
