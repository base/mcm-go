# mcmctl

CLI tool for managing MCM (Multi-Chain Multisig) on Solana.

## Installation

```bash
go build -o mcmctl ./cmd/mcmctl
```

## Configuration

### Global Flags

| Flag | Environment Variable | Description | Default |
|------|---------------------|-------------|---------|
| `--rpc, -r` | `MCM_RPC_URL` | Solana RPC endpoint URL or network alias (`mainnet`, `devnet`, `testnet`, `localhost`) | Required |
| `--ws` | `MCM_WS_URL` | WebSocket endpoint URL or network alias | Required |
| `--program-id, -p` | `MCM_PROGRAM_ID` | MCM program ID (base58) | Required |
| `--authority, -a` | `MCM_AUTHORITY` | Path to authority keypair file (also used as transaction payer) | `~/.config/solana/id.json` |

### Network Aliases

The `--rpc` and `--ws` flags support network aliases for convenience:

| Alias | RPC Endpoint | WebSocket Endpoint |
|-------|-------------|-------------------|
| `mainnet` or `mainnet-beta` | `https://api.mainnet-beta.solana.com` | `wss://api.mainnet-beta.solana.com` |
| `devnet` | `https://api.devnet.solana.com` | `wss://api.devnet.solana.com` |
| `testnet` | `https://api.testnet.solana.com` | `wss://api.testnet.solana.com` |
| `localhost` or `local` | `http://localhost:8899` | `ws://localhost:8900` |

### Example Configuration

```bash
# Using network aliases (simple)
export MCM_RPC_URL="devnet"
export MCM_WS_URL="devnet"
export MCM_PROGRAM_ID="YourProgramID"
export MCM_AUTHORITY="/path/to/authority.json"

# Using full URLs
export MCM_RPC_URL="https://api.devnet.solana.com"
export MCM_WS_URL="wss://api.devnet.solana.com"

# Or using flags with aliases
mcmctl --rpc devnet --ws devnet --program-id YourProgramID <command>

# Or using flags with full URLs
mcmctl --rpc https://api.devnet.solana.com --ws wss://api.devnet.solana.com --program-id YourProgramID <command>
```

## Commands

### Multisig Operations

#### `multisig init`

Initialize a new multisig.

```bash
mcmctl multisig init \
  --multisig-id <hex32> \
  --chain-id <uint64>
```

**Example:**

```bash
mcmctl multisig init \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --chain-id 1
```

### Signers Management

#### `signers init`

Initialize signers storage.

```bash
mcmctl signers init \
  --multisig-id <hex32> \
  --total <uint8>
```

**Example:**

```bash
mcmctl signers init \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --total 10
```

#### `signers append`

Append signers to storage.

```bash
mcmctl signers append \
  --multisig-id <hex32> \
  --signers <addr1,addr2,...> \
  [--batch-size <n>]
```

**Features:**
- Accepts EVM addresses in hex format (with or without `0x` prefix)
- Automatically sorts signers in strictly increasing order
- Validates for duplicates
- Batches signers for efficient on-chain submission

**Example:**

```bash
mcmctl signers append \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signers "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222" \
  --batch-size 10
```

#### `signers finalize`

Finalize signers (no more additions allowed).

```bash
mcmctl signers finalize \
  --multisig-id <hex32>
```

**Example:**

```bash
mcmctl signers finalize \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000
```

## Complete Workflow Example

```bash
# 1. Set environment variables (using network alias for simplicity)
export MCM_RPC_URL="devnet"
export MCM_WS_URL="devnet"
export MCM_AUTHORITY="~/.config/solana/id.json"

# 2. Initialize multisig
mcmctl multisig init \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --chain-id 1

# 3. Initialize signers storage
mcmctl signers init \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --total 3

# 4. Append signers
mcmctl signers append \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signers "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222,0x3333333333333333333333333333333333333333"

# 5. Finalize signers
mcmctl signers finalize \
  --multisig-id 6d792d6d756c74697369672d303031000000000000000000000000000000000000
```

## Error Handling

The CLI provides clear error messages:

- Invalid hex format: `Error: invalid multisig-id: invalid hex string`
- Duplicate signers: `Error: failed to parse signers: duplicate signer address`
- Missing configuration: `Error: failed to load config: RPC URL is required`

## Notes

- The authority keypair is used both as the transaction payer and the authority for all operations
- To use a different authority, specify it via `--authority` or `MCM_AUTHORITY` environment variable
- Multisig IDs must be exactly 32 bytes (64 hex characters)
- EVM signer addresses must be exactly 20 bytes (40 hex characters)
- Batch size for appending signers is limited to 1-32 per transaction
