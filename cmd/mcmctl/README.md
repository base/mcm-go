# mcmctl

CLI tool for managing MCM (Multi-Chain Multisig) on Solana.

## Installation

```bash
go build -o mcmctl ./cmd/mcmctl
```

## Configuration

### Transaction Flags (for `multisig`, `signers`, and `signatures` commands)

These flags are required for commands that interact with the Solana blockchain:

| Flag               | Environment Variable | Description                                                                            | Default                    |
| ------------------ | -------------------- | -------------------------------------------------------------------------------------- | -------------------------- |
| `--rpc, -r`        | `MCM_RPC_URL`        | Solana RPC endpoint URL or network alias (`mainnet`, `devnet`, `testnet`, `localhost`) | Required                   |
| `--ws`             | `MCM_WS_URL`         | WebSocket endpoint URL or network alias                                                | Required                   |
| `--program-id, -p` | `MCM_PROGRAM_ID`     | MCM program ID (base58)                                                                | Required                   |
| `--authority, -a`  | `MCM_AUTHORITY`      | Path to authority keypair file (also used as transaction payer)                        | `~/.config/solana/id.json` |

**Note:** `proposal` commands that work offline (like `proposal hash`) do not require these flags.

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

**Advanced Example (3 signers in hierarchical groups):**

```bash
# Group structure:
#   Group 0 (root): 2-of-3 quorum
#   Group 1: 1-of-2 quorum, parent = Group 0
#   Group 2: 2-of-2 quorum, parent = Group 0
# Signers: 2 in Group 1, 2 in Group 2, 1 in Group 0

mcmctl signers set-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --signer-groups 1,1,2,2,0 \
  --group-quorums 2,1,2 \
  --group-parents 0,0,0
```

### Proposal Operations

#### `proposal hash`

Compute the hash to sign for a proposal (offline operation).

```bash
mcmctl proposal hash --file <path>
```

**Features:**

- Works completely offline (no blockchain connection needed)
- Computes the Merkle root from the proposal
- Displays the hash to sign: `keccak256(root || validUntil)`
- Shows proposal metadata for verification

**Example:**

```bash
mcmctl proposal hash --file my_proposal.json
```

**Example Output:**

```
Proposal loaded successfully
  Multisig ID: 0x0000000000000000000000000000000000000000000000000000000000000000
  Valid Until: 1800000000
  Instructions: 1
  Chain ID: 0
  Pre Op Count: 0
  Post Op Count: 1
  Override Previous Root: false

Merkle Root: 0x92cb93881a2ea83145908fe515b581751f900d5f8f61f9c026fdc8cebd5e402c

Hash to Sign (keccak256(root || validUntil)):
vvvvvvvv
0x5be3ca56ef2891fb5f2aecbf0f826664ec0db9eddfe3cf5103cf3a9f4fc7acdb
^^^^^^^^
```

#### `proposal set-root`

Load a proposal and set its Merkle root on-chain.

```bash
mcmctl proposal set-root --file <path>
```

**Features:**

- Loads proposal from JSON file
- Computes Merkle root and metadata proof
- Submits SetRoot transaction to Solana
- Requires transaction flags (--rpc, --ws, --program-id, --authority)

**Example:**

```bash
mcmctl proposal set-root \
  --file my_proposal.json
```

**Example Output:**

```
Root set successfully
  Multisig ID: 0x0000000000000000000000000000000000000000000000000000000000000000
  Root: 0x92cb93881a2ea83145908fe515b581751f900d5f8f61f9c026fdc8cebd5e402c
  Valid Until: 1800000000
signature: 5j7s...
```

#### `proposal execute`

Execute operations from a proposal.

```bash
mcmctl proposal execute --file <path> [--operation-index <index>]
```

**Features:**

- Loads proposal from JSON file
- Computes Merkle root and operation proofs
- Executes a single operation or all operations sequentially
- Requires transaction flags (--rpc, --ws, --program-id, --authority)

**Example (execute single operation):**

```bash
mcmctl proposal execute \
  --file my_proposal.json \
  --operation-index 0
```

**Example (execute all operations):**

```bash
mcmctl proposal execute \
  --file my_proposal.json
```

**Example Output (single operation):**

```
Operation 0 executed successfully
signature: 5j7s...
```

**Example Output (all operations):**

```
Executed 3 operation(s) successfully
  operation 0: 5j7s...
  operation 1: 2k8f...
  operation 2: 9m3n...
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
  --file <path> \
  --total <uint8>
```

**Example:**

```bash
mcmctl signatures init \
  --file proposal.json \
  --total 5
```

#### `signatures append`

Append ECDSA signatures to storage.

```bash
mcmctl signatures append \
  --file <path> \
  --signatures <sig1,sig2,...>
```

**Signature Format:** Each signature must be 65 bytes in hex format (130 hex characters, must have `0x` prefix). The format is:

- Bytes 0-31: `r` value (32 bytes)
- Bytes 32-63: `s` value (32 bytes)
- Byte 64: `v` recovery ID (1 byte)

**Example (single signature):**

```bash
mcmctl signatures append \
  --file proposal.json \
  --signatures "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeffedcba0987654321fedcba0987654321fedcba0987654321fedcba09876543211b"
```

**Example (multiple signatures):**

```bash
mcmctl signatures append \
  --file proposal.json \
  --signatures "0x1234...1b,0x5678...1c"
```

#### `signatures finalize`

Finalize signatures (no more additions allowed).

```bash
mcmctl signatures finalize --file <path>
```

**Example:**

```bash
mcmctl signatures finalize --file proposal.json
```

#### `signatures clear`

Clear signature storage (allows reinitializing with different parameters).

```bash
mcmctl signatures clear --file <path>
```

**Example:**

```bash
mcmctl signatures clear --file proposal.json
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
