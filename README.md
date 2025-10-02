# MCM Go SDK

A comprehensive Go SDK for interacting with the Multi-Chain Multisig (MCM) program on Solana.

## Features

- 🔑 **PDA Derivation**: Utilities for deriving all MCM program PDAs
- 🌳 **Merkle Trees**: Keccak256-based Merkle tree construction with proof generation
- 📝 **Proposal Management**: Build and validate proposals with automatic Merkle proof generation
- 🔐 **Signature Handling**: ECDSA signature encoding/decoding
- 📡 **RPC Client**: Solana RPC client wrapper with MCM-specific functionality
- 🛠️ **High-level Services**: Simplified workflows for common operations
- 🧪 **Transaction Building**: Flexible transaction construction and submission

## Installation

```bash
go get mcm-go
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/gagliardetto/solana-go"
    "mcm-go"
    "mcm-go/pkg/client"
)

func main() {
    // Create client
    cfg := client.Config{
        RPCURL:    "https://api.devnet.solana.com",
        ProgramID: solana.MustPublicKeyFromBase58("55CNTEUq6cAa2sBA7bkDfJ2bb3uWs7Zh77vAF9H8TnJL"),
    }

    mcmClient, _ := mcm.NewClient(cfg)
    defer mcmClient.Close()

    // Create services
    proposalSvc := mcm.NewProposalService(mcmClient)

    // Build and submit proposals...
}
```

## Package Structure

```
mcm-go/
├── bindings/           # Auto-generated Anchor bindings
├── pkg/
│   ├── client/        # RPC client wrapper
│   ├── pda/           # PDA derivation utilities
│   ├── crypto/        # Merkle tree & signatures
│   ├── proposal/      # Proposal types and builders
│   ├── instructions/  # Instruction wrappers
│   ├── state/         # Account fetchers
│   ├── services/      # High-level services
│   └── tx/            # Transaction builder
├── examples/          # Usage examples
└── mcm.go            # Main SDK entry point
```

## Core Concepts

### 1. PDA Derivation

Derive Program Derived Addresses for MCM accounts:

```go
import "mcm-go/pkg/pda"

multisigID := [32]byte{...}
configPDA, _, _ := pda.MultisigConfigPDA(programID, multisigID)
```

### 2. Building Proposals

Create proposals with Merkle proofs:

```go
import "mcm-go/pkg/proposal"

builder := proposal.NewBuilder(multisigID, validUntil)
builder.SetRootMetadata(metadata)
builder.AddInstruction(instruction)

proposalWithRoot, _ := builder.Build()
```

### 3. Merkle Tree Construction

Build Merkle trees with automatic proof generation:

```go
import "mcm-go/pkg/crypto"

leaves := [][32]byte{...}
tree, _ := crypto.BuildMerkleTreeFromLeaves(leaves)

// tree.Root contains the Merkle root
// tree.Proofs[i] contains the proof for leaves[i]
```

### 4. Services

Use high-level services for common workflows:

```go
// Signers management
signersService := mcm.NewSignersService(client)
signersService.InitSigners(ctx, params)
signersService.AppendSignersInBatches(ctx, multisigID, signers, 10, authority)

// Signatures management
signaturesService := mcm.NewSignaturesService(client)
signaturesService.InitSignatures(ctx, params)
signaturesService.AppendSignaturesInBatches(ctx, multisigID, root, validUntil, sigs, 5, authority)

// Proposal operations
proposalService := mcm.NewProposalService(client)
proposalService.SetRoot(ctx, params)

// Execution
executionService := mcm.NewExecutionService(client)
executionService.ExecuteOperation(ctx, params)
```

## Workflow Example

### 1. Initialize Multisig

```go
ix, _ := mcm.Initialize(instructions.InitializeParams{
    ChainID:    1,
    MultisigID: multisigID,
    Authority:  authority,
    ProgramID:  programID,
})
```

### 2. Configure Signers

```go
// Initialize signer storage
signersService.InitSigners(ctx, InitSignersParams{
    MultisigID:   multisigID,
    TotalSigners: 10,
    Authority:    authority,
})

// Add signers in batches
signersService.AppendSignersInBatches(ctx, multisigID, signerAddresses, 5, authority)

// Finalize
signersService.FinalizeSigners(ctx, FinalizeSignersParams{
    MultisigID: multisigID,
    Authority:  authority,
})
```

### 3. Set Configuration

```go
ix, _ := mcm.SetConfig(instructions.SetConfigParams{
    MultisigID:   multisigID,
    SignerGroups: groups,
    GroupQuorums: quorums,
    GroupParents: parents,
    ClearRoot:    false,
    Authority:    authority,
    ProgramID:    programID,
})
```

### 4. Create and Sign Proposal

```go
// Build proposal
proposalToSign, _ := proposalService.CreateProposalToSign(
    multisigID,
    validUntil,
    instructions,
    metadata,
)

// Signers sign proposalToSign.HashToSign off-chain
// Collect ECDSA signatures...

// Submit signatures
signaturesService.InitSignatures(ctx, params)
signaturesService.AppendSignaturesInBatches(ctx, multisigID, root, validUntil, signatures, 5, authority)
signaturesService.FinalizeSignatures(ctx, params)
```

### 5. Set Root

```go
proposalService.SetRoot(ctx, SetRootParams{
    MultisigID: multisigID,
    Proposal:   proposalWithRoot,
    Authority:  authority,
})
```

### 6. Execute Operations

```go
executionService.ExecuteAllOperations(
    ctx,
    multisigID,
    proposalWithRoot,
    authority,
)
```

## Account Fetching

Fetch on-chain account state:

```go
import "mcm-go/pkg/state"

fetcher := state.NewFetcher(rpcClient, programID)

// Fetch individual accounts
config, _ := fetcher.GetMultisigConfig(ctx, multisigID)
metadata, _ := fetcher.GetRootMetadata(ctx, multisigID)

// Fetch all state at once
state, _ := fetcher.GetMultisigState(ctx, multisigID)
```

## Testing

```bash
go test ./...
```

## Architecture

The SDK follows a layered architecture:

1. **Bindings Layer** (`bindings/`): Auto-generated Anchor code
2. **Core Layer** (`pkg/pda`, `pkg/crypto`): Fundamental utilities
3. **Proposal Layer** (`pkg/proposal`): Proposal construction
4. **Instruction Layer** (`pkg/instructions`): Instruction builders
5. **State Layer** (`pkg/state`): Account fetching
6. **Service Layer** (`pkg/services`): High-level workflows
7. **Client Layer** (`pkg/client`, `pkg/tx`): RPC and transaction handling

## Contributing

See [MIGRATION.md](./MIGRATION.md) for implementation details and architecture decisions.

## License

MIT

## Links

- [MCM Program Documentation](https://github.com/smartcontractkit/chainlink-mcm)
- [Solana Documentation](https://docs.solana.com/)
- [TypeScript SDK Reference](./mcm-sol-ts/)
