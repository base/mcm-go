# mcmctl

CLI tool for managing MCM (Multi-Chain Multisig) on Solana.

## Installation

```bash
go install github.com/base/mcm-go/cmd/mcmctl@latest
```

## Configuration

### Flags Overview

Commands require different sets of flags depending on their operation:

**Read-only commands** (e.g., `proposal create`):

- `--rpc, -r` : RPC endpoint (required)
- `--program-id, -p` : MCM program ID (required)

**Write commands** (e.g., `multisig init`, `signers *`, `proposal set-root`, `proposal execute`):

- `--rpc, -r` : RPC endpoint (required)
- `--program-id, -p` : MCM program ID (required)
- `--ws` : WebSocket endpoint (required for confirmations)
- `--authority, -a` : Keypair file for signing (defaults to `~/.config/solana/id.json`)

**Offline commands** (e.g., `proposal hash`):

- No blockchain flags needed

### Flag Details

| Flag               | Environment Variable | Description                                                                            | Default                                    |
| ------------------ | -------------------- | -------------------------------------------------------------------------------------- | ------------------------------------------ |
| `--rpc, -r`        | `MCM_RPC_URL`        | Solana RPC endpoint URL or network alias (`mainnet`, `devnet`, `testnet`, `localhost`) | Required (for on-chain ops)                |
| `--ws`             | `MCM_WS_URL`         | WebSocket endpoint URL or network alias                                                | Required (for write ops)                   |
| `--program-id, -p` | `MCM_PROGRAM_ID`     | MCM program ID (base58)                                                                | Required (for on-chain ops)                |
| `--authority, -a`  | `MCM_AUTHORITY`      | Path to authority keypair file (also used as transaction payer)                        | `~/.config/solana/id.json` (for write ops) |

### Network Aliases

The `--rpc` and `--ws` flags support network aliases for convenience:

| Alias                       | RPC Endpoint                          | WebSocket Endpoint                  |
| --------------------------- | ------------------------------------- | ----------------------------------- |
| `mainnet` or `mainnet-beta` | `https://api.mainnet-beta.solana.com` | `wss://api.mainnet-beta.solana.com` |
| `devnet`                    | `https://api.devnet.solana.com`       | `wss://api.devnet.solana.com`       |
| `testnet`                   | `https://api.testnet.solana.com`      | `wss://api.testnet.solana.com`      |
| `localhost` or `local`      | `http://localhost:8899`               | `ws://localhost:8900`               |

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
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --chain-id 1
```

#### `multisig print-authority`

Print the multisig signer PDA (authority address).

```bash
mcmctl multisig print-authority \
  --multisig-id <hex32>
```

**Example:**

```bash
mcmctl multisig print-authority \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
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
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --total 10
```

#### `signers append`

Append signers to storage.

```bash
mcmctl signers append \
  --multisig-id <hex32> \
  --signers <addr1,addr2,...>
```

**Features:**

- Accepts EVM addresses in hex format (must have `0x` prefix)
- Automatically sorts signers in strictly increasing order
- Validates for duplicates

**Example:**

```bash
mcmctl signers append \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signers "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222"
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
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
```

#### `signers set-config`

Set the multisig configuration (signer groups and quorums).

```bash
mcmctl signers set-config \
  --multisig-id <hex32> \
  --signer-groups <group0,group1,...> \
  --group-quorums <quorum0,quorum1,...> \
  --group-parents <parent0,parent1,...> \
  [--clear-root]
```

**Features:**

- Automatically pads `--group-quorums` and `--group-parents` to 32 elements with zeros
- Validates signer group hierarchy and quorum requirements
- Optional `--clear-root` flag to invalidate the current Merkle root

**Simple Example (1 signer in root group):**

```bash
mcmctl signers set-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signer-groups 0 \
  --group-quorums 1 \
  --group-parents 0 \
  --clear-root
```

**Advanced Example (5 signers in hierarchical groups):**

```bash
# Group structure:
#   Group 0 (root): Has 1 direct signer + 2 child groups = 3 entities
#                   Quorum: 2-of-3 (need 2 of: direct signer, Group 1, Group 2)
#   Group 1: 2 signers, quorum 1-of-2, parent = Group 0
#   Group 2: 2 signers, quorum 2-of-2, parent = Group 0

mcmctl signers set-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signer-groups 1,1,2,2,0 \
  --group-quorums 2,1,2 \
  --group-parents 0,0,0
```

### Proposal Operations

#### `proposal create`

Create a complete proposal by combining instructions from a JSON file with on-chain state.

```bash
mcmctl proposal create \
  --instructions <path> \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Loads instructions from a simplified JSON file (only operations, no metadata)
- Fetches current on-chain state (chain ID, multisig, op counts) automatically
- Generates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc` and `--program-id`, no wallet needed)

**Instructions JSON Format:**

The input file contains only the instructions array:

```json
[
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
  }
]
```

**Example:**

```bash
# Create a proposal from instructions
mcmctl proposal create \
  --instructions ./instructions.json \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./proposal.json
```

The generated `proposal.json` file contains the complete proposal with all metadata and can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal hash`

Compute the hash to sign for a proposal (offline operation).

```bash
mcmctl proposal hash --proposal <path>
```

**Features:**

- Works completely offline (no blockchain connection needed)
- Computes the Merkle root from the proposal
- Displays the hash to sign: `keccak256(root || validUntil)`
- Shows proposal metadata for verification

**Example:**

```bash
mcmctl proposal hash --proposal my_proposal.json
```

#### `proposal set-root`

Load a proposal and set its Merkle root on-chain.

```bash
mcmctl proposal set-root --proposal <path>
```

**Features:**

- Loads proposal from JSON file
- Computes Merkle root and metadata proof
- Submits SetRoot transaction to Solana
- Requires transaction flags (--rpc, --ws, --program-id, --authority)

**Example:**

```bash
mcmctl proposal set-root \
  --proposal my_proposal.json
```

#### `proposal execute`

Execute operations from a proposal.

```bash
mcmctl proposal execute --proposal <path> [--operation-count <count>] [--start-index <index>]
```

**Flags:**

- `--operation-count, -n` : Number of operations to execute (defaults to all remaining operations)
- `--start-index, -s` : Index of first operation to execute (default: 0)

**Features:**

- Loads proposal from JSON file
- Computes Merkle root and operation proofs
- Executes operations sequentially starting from start-index
- If `--operation-count` is not specified, executes all remaining operations from start-index
- Requires transaction flags (--rpc, --ws, --program-id, --authority)

**Example (execute all operations):**

```bash
mcmctl proposal execute \
  --proposal my_proposal.json
```

**Example (execute single operation):**

```bash
mcmctl proposal execute \
  --proposal my_proposal.json \
  --operation-count 1
```

**Example (execute operations 1 and 2):**

```bash
mcmctl proposal execute \
  --proposal my_proposal.json \
  --start-index 1 \
  --operation-count 2
```

**Example (execute all remaining operations starting from index 2):**

```bash
mcmctl proposal execute \
  --proposal my_proposal.json \
  --start-index 2
```

**Proposal JSON Format:**

```json
{
  "multisigId": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "validUntil": 1800000000,
  "rootMetadata": {
    "chainId": 0,
    "multisig": "11111111111111111111111111111112",
    "preOpCount": 0,
    "postOpCount": 1,
    "overridePreviousRoot": false
  },
  "instructions": [
    {
      "programId": "11111111111111111111111111111111",
      "accounts": [
        {
          "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
          "isSigner": true,
          "isWritable": false
        }
      ],
      "data": "0xdeadbeef"
    }
  ]
}
```

### Signature Management

These commands manage ECDSA signatures for setting new Merkle roots.

#### `signatures init`

Initialize signature storage for a new root.

```bash
mcmctl signatures init \
  --proposal <path> \
  --total <uint8>
```

**Example:**

```bash
mcmctl signatures init \
  --proposal proposal.json \
  --total 5
```

#### `signatures append`

Append ECDSA signatures to storage.

```bash
mcmctl signatures append \
  --proposal <path> \
  --signatures <sig1,sig2,...>
```

**Signature Format:** Each signature must be 65 bytes in hex format (130 hex characters, must have `0x` prefix). The format is:

- Bytes 0-31: `r` value (32 bytes)
- Bytes 32-63: `s` value (32 bytes)
- Byte 64: `v` recovery ID (1 byte)

**Example (single signature):**

```bash
mcmctl signatures append \
  --proposal proposal.json \
  --signatures "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeffedcba0987654321fedcba0987654321fedcba0987654321fedcba09876543211b"
```

**Example (multiple signatures):**

```bash
mcmctl signatures append \
  --proposal proposal.json \
  --signatures "0x1234...1b,0x5678...1c"
```

#### `signatures finalize`

Finalize signatures (no more additions allowed).

```bash
mcmctl signatures finalize --proposal <path>
```

**Example:**

```bash
mcmctl signatures finalize --proposal proposal.json
```

#### `signatures clear`

Clear signature storage (allows reinitializing with different parameters).

```bash
mcmctl signatures clear --proposal <path>
```

**Example:**

```bash
mcmctl signatures clear --proposal proposal.json
```

## Complete Workflow Example

```bash
# 1. Set environment variables (using network alias for simplicity)
export MCM_RPC_URL="devnet"
export MCM_WS_URL="devnet"
export MCM_AUTHORITY="~/.config/solana/id.json"
export MCM_PROGRAM_ID="YourProgramID"

# 2. Initialize multisig
mcmctl multisig init \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --chain-id 1

# 3. Initialize signers storage
mcmctl signers init \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --total 3

# 4. Append signers
mcmctl signers append \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signers "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222,0x3333333333333333333333333333333333333333"

# 5. Finalize signers
mcmctl signers finalize \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000

# 6. Set configuration (simple: all 3 signers in root group, 2-of-3 quorum)
mcmctl signers set-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signer-groups 0,0,0 \
  --group-quorums 2 \
  --group-parents 0 \
  --clear-root
```
