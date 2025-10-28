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

- `--rpc-url` : RPC endpoint (required)
- `--mcm-program-id` : MCM program ID (required)

**Write commands** (e.g., `multisig init`, `signers *`, `proposal set-root`, `proposal execute`):

- `--rpc-url` : RPC endpoint (required)
- `--mcm-program-id` : MCM program ID (required)
- `--ws-url` : WebSocket endpoint (required for confirmations)
- `--authority` : Keypair file for signing (defaults to `~/.config/solana/id.json`)

**Offline commands** (e.g., `proposal hash`):

- No blockchain flags needed

### Flag Details

| Flag               | Environment Variable | Description                                                                            | Default                                    |
| ------------------ | -------------------- | -------------------------------------------------------------------------------------- | ------------------------------------------ |
| `--rpc-url`        | `RPC_URL`            | Solana RPC endpoint URL or network alias (`mainnet`, `devnet`, `testnet`, `localhost`) | Required (for on-chain ops)                |
| `--ws-url`         | `WS_URL`             | WebSocket endpoint URL or network alias                                                | Required (for write ops)                   |
| `--mcm-program-id` | `MCM_PROGRAM_ID`     | MCM program ID (base58)                                                                | Required (for on-chain ops)                |
| `--authority`      | `MCM_AUTHORITY`      | Path to authority keypair file (also used as transaction payer)                        | `~/.config/solana/id.json` (for write ops) |

### Network Aliases

The `--rpc-url` and `--ws-url` flags support network aliases for convenience:

| Alias                       | RPC Endpoint                          | WebSocket Endpoint                  |
| --------------------------- | ------------------------------------- | ----------------------------------- |
| `mainnet` or `mainnet-beta` | `https://api.mainnet-beta.solana.com` | `wss://api.mainnet-beta.solana.com` |
| `devnet`                    | `https://api.devnet.solana.com`       | `wss://api.devnet.solana.com`       |
| `testnet`                   | `https://api.testnet.solana.com`      | `wss://api.testnet.solana.com`      |
| `localhost` or `local`      | `http://localhost:8899`               | `ws://localhost:8900`               |

### Example Configuration

```bash
# Using network aliases (simple)
export RPC_URL="devnet"
export WS_URL="devnet"
export MCM_PROGRAM_ID="YourProgramID"
export MCM_AUTHORITY="/path/to/authority.json"

# Using full URLs
export RPC_URL="https://api.devnet.solana.com"
export WS_URL="wss://api.devnet.solana.com"

# Or using flags with aliases
mcmctl --rpc-url devnet --ws-url devnet --mcm-program-id YourProgramID <command>

# Or using flags with full URLs
mcmctl --rpc-url https://api.devnet.solana.com --ws-url wss://api.devnet.solana.com --mcm-program-id YourProgramID <command>
```

## Commands

### Multisig Operations

#### `multisig init`

Initialize a new multisig on-chain.

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

### Ownership Management

Ownership commands allow secure transfer of multisig ownership using a two-step process to prevent accidental transfers.

#### `ownership transfer`

Propose a new owner for the multisig (step 1 of 2).

```bash
mcmctl ownership transfer \
  --multisig-id <hex32> \
  --proposed-owner <base58_pubkey>
```

**Features:**

- Only the current owner can propose a new owner
- Proposed owner must be a valid Solana public key (base58 format)
- Proposed owner cannot be the same as the current owner
- This is step 1 of a secure two-step ownership transfer process

**Example:**

```bash
mcmctl ownership transfer \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --proposed-owner 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
```

**Output:**

```
ownership transfer proposed
proposed owner: 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
signature: <transaction_signature>
```

#### `ownership accept`

Accept ownership of the multisig (step 2 of 2).

```bash
mcmctl ownership accept \
  --multisig-id <hex32>
```

**Features:**

- Only the proposed owner can accept ownership
- The `--authority` flag must point to the proposed owner's keypair
- Once accepted, ownership transfer is complete and permanent
- The `proposed_owner` field is reset after acceptance

**Example:**

```bash
# Use the proposed owner's keypair as authority
mcmctl ownership accept \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --authority ~/.config/solana/new-owner-keypair.json
```

**Output:**

```
ownership accepted
new owner: 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
signature: <transaction_signature>
```

**Important Notes:**

- **Two-step process**: Transfer must be proposed by current owner AND accepted by new owner
- **Proof of control**: New owner must prove they control the private key by signing the acceptance transaction
- **No rollback**: Once accepted, the transfer is permanent (unless a new transfer is initiated)
- **Security**: This pattern prevents accidental transfers to incorrect addresses

### Signers Management

#### `signers print-config`

Display the current multisig configuration including signers, groups, and quorums.

```bash
mcmctl signers print-config \
  --multisig-id <hex32> \
  [--pretty]
```

**Flags:**

- `--pretty`: Display configuration as a tree hierarchy (optional)

**Example (flat format):**

```bash
mcmctl signers print-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
```

**Output:**

```
=== Multisig Configuration ===
Multisig ID: 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
Chain ID: 1
Owner: 9aXm5rpgGKaL...
Proposed Owner: 11111111111111111111111111111111

=== Signers (3 total) ===
  [0] Address: 0x1234..., Index: 0, Group: 1
  [1] Address: 0x5678..., Index: 1, Group: 1
  [2] Address: 0xabcd..., Index: 2, Group: 2

=== Group Quorums ===
  Group 0: quorum = 2
  Group 1: quorum = 2
  Group 2: quorum = 1

=== Group Parents (Hierarchy) ===
  Group 0: ROOT (no parent)
  Group 1: parent = Group 0
  Group 2: parent = Group 0
```

**Example (pretty format):**

```bash
mcmctl signers print-config \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --pretty
```

**Output:**

```
=== Multisig Configuration ===
Multisig ID: 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
Chain ID: 1
Owner: 9aXm5rpgGKaL...
Proposed Owner: 11111111111111111111111111111111

=== Group Hierarchy ===
Group 0 (ROOT, quorum: 2)
├── Group 1 (quorum: 2)
│   ├── [Signer 0] 0x1234...
│   └── [Signer 1] 0x5678...
└── Group 2 (quorum: 1)
    └── [Signer 2] 0xabcd...
```

**Features:**

- Read-only operation (only requires `--rpc-url` and `--program-id`, no wallet needed)
- Tree visualization with `--pretty` flag for easy understanding of group hierarchy
- Displays complete configuration state
- Shows signer addresses with their EVM addresses, indices, and group assignments
- Shows group quorum requirements
- Shows group hierarchy (parent-child relationships)

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
- Automatically batches signers in chunks of 10 per transaction (sends multiple transactions if needed)

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

#### `signers clear`

Clear signers storage.

```bash
mcmctl signers clear \
  --multisig-id <hex32>
```

**Example:**

```bash
mcmctl signers clear \
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
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

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

#### `proposal loader-v3 upgrade`

Create a proposal to upgrade a Solana program using BPF Loader v3.

```bash
mcmctl proposal loader-v3 upgrade \
  --program <base58_pubkey> \
  --buffer <base58_pubkey> \
  --spill <base58_pubkey> \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Automatically derives the ProgramData PDA from the program address
- Fetches and validates upgrade authorities from on-chain (program and buffer must match)
- Creates the BPF Loader v3 upgrade instruction
- Generates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

**Example:**

```bash
# Create an upgrade proposal
mcmctl proposal loader-v3 upgrade \
  --program 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde \
  --buffer BUFFERAccountXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --spill SPILLAccountXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./upgrade_proposal.json
```

**Output:**

```
ProgramData (derived): AbCdEf...
Upgrade authority: 9aE476sH92...
Buffer authority: 9aE476sH92...
Fetching MCM on-chain state...

Upgrade proposal created successfully and saved to ./upgrade_proposal.json
  Program: 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
  Buffer: BUFFERAccountXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  Spill: SPILLAccountXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  Multisig ID: 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
  Valid Until: 1800000000
  Chain ID: 1
  Pre Op Count: 0
  Post Op Count: 1
```

**Workflow:**

1. **Prepare buffer**: Use `solana program write-buffer` to upload your new program
2. **Create proposal**: Use this command to generate the upgrade proposal
3. **Sign proposal**: Use `proposal hash` to get the hash, collect signatures
4. **Submit signatures**: Use `signatures init/append/finalize` to submit signatures
5. **Set root**: Use `proposal set-root` to set the Merkle root on-chain
6. **Execute**: Use `proposal execute` to perform the upgrade

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal loader-v3 set-authority`

Create a proposal to change or remove the upgrade authority of a BPF Loader v3 program or buffer.

```bash
mcmctl proposal loader-v3 set-authority \
  --account <program_id_or_buffer> \
  --new-authority <base58_pubkey> \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Automatically detects if account is a Program ID or Buffer address
- If Program ID: derives ProgramData PDA automatically
- If Buffer: uses the address directly
- Omitting `--new-authority` makes the program/buffer **immutable** (irreversible)
- Read-only operation (no wallet needed)

**Example (change authority):**

```bash
mcmctl proposal loader-v3 set-authority \
  --account 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde \
  --new-authority NewAuthorityXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./set_authority_proposal.json
```

**Example (make immutable - no `--new-authority`):**

```bash
mcmctl proposal loader-v3 set-authority \
  --account 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./make_immutable_proposal.json
```

**⚠️ Warning:** Making a program or buffer immutable is **irreversible**. No further upgrades or authority changes will be possible.

**Workflow:**

1. **Create proposal**: Use this command
2. **Sign proposal**: Use `proposal hash` to get the hash, collect signatures
3. **Submit signatures**: Use `signatures init/append/finalize` to submit signatures
4. **Set root**: Use `proposal set-root` to set the Merkle root on-chain
5. **Execute**: Use `proposal execute` to perform the authority change

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal mcm update-signers`

Create a proposal to update MCM signers configuration with a complete workflow (init, append, finalize, setConfig).

```bash
mcmctl proposal mcm update-signers \
  --new-signers <comma_separated_hex_addresses> \
  --signer-groups <comma_separated_group_indices> \
  --group-quorums <comma_separated_quorums> \
  --group-parents <comma_separated_parents> \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--clear-root] \
  [--clear-signers] \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Generates a complete signers update workflow in a single proposal:
  0. **ClearSigners** (optional): Clears existing signers if `--clear-signers` flag is set
  1. **InitSigners**: Initializes signer storage with total count
  2. **AppendSigners**: Adds all signers in chunks (max 10 per instruction)
  3. **FinalizeSigners**: Finalizes the signer list
  4. **SetConfig**: Configures signer groups, quorums, and parent relationships
- Automatically sorts signers in strictly increasing order (warns if reordering occurs)
- Automatically derives the MCM authority PDA
- Creates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

**Parameters:**

- `--new-signers`: Comma-separated list of 20-byte hex addresses (with 0x prefix). Automatically sorted in strictly increasing order.
- `--signer-groups`: Group index for each signer (e.g., "0,0,1,1" for 4 signers in 2 groups)
- `--group-quorums`: Quorum threshold for each group (e.g., "2,3" for 2 groups)
- `--group-parents`: Parent group index for each group (e.g., "0,0" for 2 root groups)
- `--clear-root`: Optional flag to clear existing Merkle root (invalidates pending operations)
- `--clear-signers`: Optional flag to clear existing signers before updating (adds ClearSigners instruction at the beginning)

**Example:**

```bash
# Create a signers update proposal with 4 signers in 2 groups
mcmctl proposal mcm update-signers \
  --new-signers 0x1234567890123456789012345678901234567890,0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x9876543210987654321098765432109876543210,0xfedcbafedcbafedcbafedcbafedcbafedcbafedc \
  --signer-groups 0,0,1,1 \
  --group-quorums 2,2 \
  --group-parents 0,0 \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./update_signers_proposal.json
```

**Output:**

```
New signers: 4
  [0] 0x1234567890123456789012345678901234567890
  [1] 0xabcdefabcdefabcdefabcdefabcdefabcdefabcd
  [2] 0x9876543210987654321098765432109876543210
  [3] 0xfedcbafedcbafedcbafedcbafedcbafedcbafedc
Signer groups: [0 0 1 1]
Group quorums: [2 2 0 0...]
Group parents: [0 0 0 0...]
Clear root: false
MCM authority: AbCdEf...

1. Creating InitSigners instruction...
   ✓ Total signers: 4

2. Creating AppendSigners instruction(s)...
   ✓ Chunk 1: 4 signers [0:4]

3. Creating FinalizeSigners instruction...
   ✓ Finalize complete

4. Creating SetConfig instruction...
   ✓ Groups: 2, ClearRoot: false

Fetching MCM on-chain state...

✅ Update signers proposal created successfully and saved to ./update_signers_proposal.json
  Multisig ID: 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000
  Valid Until: 1800000000
  Instructions: 4
  Chain ID: 1
  Pre Op Count: 0
  Post Op Count: 4
```

**Workflow:**

This command generates all 4 instructions needed to completely update signers:

1. Initialize storage for new signers
2. Append all signer addresses (in chunks if needed)
3. Finalize the signer list
4. Set the configuration (groups, quorums, hierarchy)

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal mcm accept-ownership`

Create a proposal to accept ownership of a program or account.

```bash
mcmctl proposal mcm accept-ownership \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Creates a proposal to accept ownership (second step of ownership transfer)
- Fetches current on-chain state automatically
- Generates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

**Example:**

```bash
mcmctl proposal mcm accept-ownership \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./accept_ownership_proposal.json
```

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal bridge pause`

Create a proposal to pause/unpause bridge operations.

```bash
mcmctl proposal bridge pause \
  --bridge-program-id <base58_pubkey> \
  (--pause | --unpause) \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Creates a proposal to set the pause status of the bridge
- The bridge account PDA is automatically derived from the bridge program ID (using seed `"bridge"`)
- The guardian is automatically fetched from the Bridge account on-chain
- Use `--pause` to pause bridge operations or `--unpause` to unpause (mutually exclusive)
- Fetches current on-chain state automatically
- Generates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

**Example:**

```bash
# Pause the bridge
mcmctl proposal bridge pause \
  --bridge-program-id BridgeProgramIDXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --pause \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./pause_bridge_proposal.json

# Unpause the bridge
mcmctl proposal bridge pause \
  --bridge-program-id BridgeProgramIDXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --unpause \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./unpause_bridge_proposal.json
```

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal bridge set-partner-oracle-config`

Create a proposal to update the partner oracle configuration in the bridge.

```bash
mcmctl proposal bridge set-partner-oracle-config \
  --bridge-program-id <base58_pubkey> \
  --required-threshold <uint8> \
  --multisig-id <hex32> \
  --valid-until <timestamp> \
  [--override-previous-root] \
  --output <path>
```

**Features:**

- Creates a proposal to set the partner oracle configuration
- The bridge account PDA is automatically derived from the bridge program ID (using seed `"bridge"`)
- The upgrade authority is automatically fetched from the bridge program's ProgramData account
- Required threshold: number of oracle signatures required (0-255)
- Fetches current on-chain state automatically
- Generates a complete proposal file ready for signing and execution
- Read-only operation (only requires `--rpc-url` and `--mcm-program-id`, no wallet needed)

**Example:**

```bash
mcmctl proposal bridge set-partner-oracle-config \
  --bridge-program-id BridgeProgramIDXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
  --required-threshold 3 \
  --multisig-id 0x6d792d6d756c74697369672d303031000000000000000000000000000000000000 \
  --valid-until 1800000000 \
  --output ./set_partner_oracle_config_proposal.json
```

The generated proposal can be used with `proposal hash`, `proposal set-root`, and `proposal execute` commands.

#### `proposal hash`

Compute the message hash for a proposal (offline operation).

```bash
mcmctl proposal hash --proposal <path>
```

**Features:**

- Works completely offline (no blockchain connection needed)
- Computes the Merkle root from the proposal
- Displays the message hash: EIP-712 structured data hash
- Shows proposal metadata for verification

**Example:**

```bash
mcmctl proposal hash --proposal my_proposal.json
```

#### `proposal eip712`

Display the complete EIP-712 typed data payload for a proposal (offline operation).

```bash
mcmctl proposal eip712 --proposal <path> --mcm-program-id <program_id>
```

**Features:**

- Works completely offline (no blockchain connection needed)
- Outputs the complete EIP-712 JSON payload
- Compatible with external signing tools (Ledger, MetaMask, Safe, etc.)
- Shows the full typed data structure including types, domain, and message

**Example:**

```bash
mcmctl proposal eip712 \
  --proposal my_proposal.json \
  --mcm-program-id 55CNTEUq6cAa2sBA7bkDfJ2bb3uWs7Zh77vAF9H8TnJL
```

**Output:**

```json
{
  "types": {
    "EIP712Domain": [
      { "name": "name", "type": "string" },
      { "name": "version", "type": "string" },
      { "name": "chainId", "type": "uint256" },
      { "name": "salt", "type": "bytes32" }
    ],
    "RootValidation": [
      { "name": "root", "type": "bytes32" },
      { "name": "validUntil", "type": "uint32" }
    ]
  },
  "primaryType": "RootValidation",
  "domain": {
    "name": "ManyChainMultiSig",
    "version": "1",
    "chainId": "0x1",
    "salt": "0x..."
  },
  "message": {
    "root": "0x...",
    "validUntil": "1800000000"
  }
}
```

This JSON can be directly used with EIP-712 signing tools.

#### `proposal set-root`

Load a proposal and set its Merkle root on-chain.

```bash
mcmctl proposal set-root --proposal <path>
```

**Features:**

- Loads proposal from JSON file
- Computes Merkle root and metadata proof
- Submits SetRootEip712 transaction to Solana
- Requires transaction flags (--rpc-url, --ws-url, --mcm-program-id, --authority)

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
- Requires transaction flags (--rpc-url, --ws-url, --mcm-program-id, --authority)

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
export RPC_URL="devnet"
export WS_URL="devnet"
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
