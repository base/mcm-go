# MCM Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/base/mcm-go)](https://goreportcard.com/report/github.com/base/mcm-go)
[![Build Status](https://github.com/base/mcm-go/workflows/CI/badge.svg)](https://github.com/base/mcm-go/actions)
[![License](https://img.shields.io/github/license/base/mcm-go)](LICENSE)

Go SDK for the Multi-Chain Multisig (MCM) Solana program.

## Overview

This SDK provides a comprehensive interface for interacting with the MCM Solana program, enabling secure multi-signature operations across chains. It includes both programmatic Go bindings and a CLI tool for managing multisig configurations, signers, and ownership transfers.

## Features

- **Multisig Management**: Initialize and configure multi-signature accounts
- **Ownership Transfer**: Two-step ownership transfer process for security
- **Signer Administration**: Add, remove, and manage authorized signers
- **CLI Tool**: Command-line interface for common operations
- **RPC Integration**: Built-in support for Solana RPC and WebSocket connections

## Installation

### SDK

```bash
go get github.com/base/mcm-go
```

### CLI Tool

```bash
go install github.com/base/mcm-go/cmd/mcmctl@latest
```

## Configuration

Set the following environment variables:

```bash
export RPC_URL="devnet"          # RPC endpoint or network name
export WS_URL="devnet"           # WebSocket endpoint or network name
export MCM_PROGRAM_ID="<id>"     # MCM program identifier
```

## CLI Commands

### Initialize Multisig

Create a new multisig account with specified parameters:

```bash
mcmctl multisig init --multisig-id <hex32> --chain-id 1
```

*Note: Hex values must use `0x` prefix*

### Transfer Ownership

Two-step process to securely transfer multisig ownership:

```bash
# Step 1: Propose new owner
mcmctl ownership transfer --multisig-id <hex32> --proposed-owner <pubkey>

# Step 2: Accept ownership with new owner's authority
mcmctl ownership accept --multisig-id <hex32> --authority <new-owner-keypair>
```

### Manage Signers

Initialize signer configuration:

```bash
mcmctl signers init --multisig-id <hex32> --total 10
```

Add new signer:

```bash
mcmctl signers add --multisig-id <hex32> --signer <pubkey>
```

Remove existing signer:

```bash
mcmctl signers remove --multisig-id <hex32> --signer <pubkey>
```

## Documentation

- [API Reference](API.md) - Complete API documentation
- [Audit S3 Format](AUDIT_S3_FORMAT.md) - S3 audit log format specification
- [Bundle States](BUNDLE_STATES.md) - Transaction bundle state machine

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## Contributing

Contributions are welcome. Please ensure all tests pass and code follows Go conventions before submitting pull requests.

## License

See [LICENSE](LICENSE) for details.

## Changelog

*Coming soon*
