package flags

import (
	"github.com/base/mcm-go/cmd/mcmctl/util"

	ucli "github.com/urfave/cli/v2"
)

// =============================================================================
// Infrastructure Flags (RPC, WS, Program, Authority)
// =============================================================================

func MCMProgramIDFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "mcm-program-id",
		Usage:    "MCM program ID (base58)",
		EnvVars:  []string{"MCM_PROGRAM_ID"},
		Required: true,
	}
}

func RPCFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "rpc-url",
		Usage:    "Solana RPC endpoint URL",
		EnvVars:  []string{"RPC_URL"},
		Required: true,
	}
}

func WSFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "ws-url",
		Usage:    "Solana WebSocket endpoint URL (required for confirmations)",
		EnvVars:  []string{"WS_URL"},
		Required: true,
	}
}

func AuthorityFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:    "authority",
		Usage:   "Path to authority keypair file (JSON or base58, also used as transaction payer)",
		EnvVars: []string{"MCM_AUTHORITY"},
		Value:   util.DefaultKeypairPath(),
	}
}

func OnchainReadFlags() []ucli.Flag {
	return []ucli.Flag{
		MCMProgramIDFlag(),
		RPCFlag(),
	}
}

func OnchainWriteFlags() []ucli.Flag {
	return append(
		OnchainReadFlags(),
		WSFlag(),
		AuthorityFlag(),
	)
}

// =============================================================================
// Multisig Flags
// =============================================================================

func MultisigIDFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "multisig-id",
		Usage:    "Multisig identifier (32 bytes hex)",
		EnvVars:  []string{"MCM_MULTISIG_ID"},
		Required: true,
	}
}

func ChainIDFlag() ucli.Flag {
	return &ucli.Uint64Flag{
		Name:     "chain-id",
		Usage:    "Chain ID (uint64)",
		EnvVars:  []string{"MCM_CHAIN_ID"},
		Required: true,
	}
}

func ProposedOwnerFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "proposed-owner",
		Usage:    "Proposed new owner public key (base58)",
		EnvVars:  []string{"MCM_PROPOSED_OWNER"},
		Required: true,
	}
}

// =============================================================================
// Proposal Flags
// =============================================================================

func ProposalFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "proposal",
		Usage:    "Path to proposal JSON file",
		EnvVars:  []string{"MCM_PROPOSAL"},
		Required: true,
	}
}

func InstructionsFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "instructions",
		Usage:    "Path to instructions JSON file (simplified format with only instructions array)",
		Required: true,
	}
}

func ValidUntilFlag() ucli.Flag {
	return &ucli.Uint64Flag{
		Name:     "valid-until",
		Usage:    "Proposal expiration timestamp (Unix timestamp)",
		Required: true,
	}
}

func OverridePreviousRootFlag() ucli.Flag {
	return &ucli.BoolFlag{
		Name:  "override-previous-root",
		Usage: "Override previous Merkle root (invalidates pending operations)",
	}
}

func OutputFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "output",
		Usage:    "Output file path for the generated proposal",
		Required: true,
	}
}

func StartIndexFlag() ucli.Flag {
	return &ucli.UintFlag{
		Name:  "start-index",
		Usage: "Index of first operation to execute",
		Value: 0,
	}
}

func OperationCountFlag() ucli.Flag {
	return &ucli.UintFlag{
		Name:  "operation-count",
		Usage: "Number of operations to execute (defaults to all remaining operations)",
	}
}

// ProposalCreationFlags returns common flags for all concrete proposal commands
// These flags are used by all commands that create MCM proposals (loader_v3, mcm, bridge)
func ProposalCreationFlags() []ucli.Flag {
	return append(
		OnchainReadFlags(),
		MultisigIDFlag(),
		ValidUntilFlag(),
		OverridePreviousRootFlag(),
		OutputFlag(),
	)
}

// =============================================================================
// Signer Flags
// =============================================================================

func SignersFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "signers",
		Usage:    "Comma-separated list of signer addresses (must start with 0x prefix)",
		EnvVars:  []string{"MCM_SIGNERS"},
		Required: true,
	}
}

func TotalSignersFlag() ucli.Flag {
	return &ucli.IntFlag{
		Name:     "total",
		Usage:    "Total number of signers",
		EnvVars:  []string{"MCM_TOTAL_SIGNERS"},
		Required: true,
	}
}

func SignerGroupsFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "signer-groups",
		Usage:    "Comma-separated group assignment for each signer (e.g., '0,1,0' for 3 signers)",
		EnvVars:  []string{"MCM_SIGNER_GROUPS"},
		Required: true,
	}
}

func GroupQuorumsFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "group-quorums",
		Usage:    "Comma-separated quorum for each group (automatically padded to 32, e.g., '1' or '2,1,1')",
		EnvVars:  []string{"MCM_SIGNER_GROUPS_QUORUMS"},
		Required: true,
	}
}

func GroupParentsFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "group-parents",
		Usage:    "Comma-separated parent group for each group (automatically padded to 32, e.g., '0' or '0,0,1')",
		EnvVars:  []string{"MCM_SIGNER_GROUPS_PARENTS"},
		Required: true,
	}
}

func ClearRootFlag() ucli.Flag {
	return &ucli.BoolFlag{
		Name:    "clear-root",
		Usage:   "Clear the current Merkle root (invalidates pending operations)",
		EnvVars: []string{"MCM_CLEAR_ROOT"},
	}
}

// =============================================================================
// Signature Flags
// =============================================================================

func SignaturesFlag() ucli.Flag {
	return &ucli.StringFlag{
		Name:     "signatures",
		Usage:    "Comma-separated ECDSA signatures (must start with 0x, format: 0x<130 hex chars>)",
		EnvVars:  []string{"MCM_SIGNATURES"},
		Required: true,
	}
}

func TotalSignaturesFlag() ucli.Flag {
	return &ucli.IntFlag{
		Name:     "total",
		Usage:    "Total number of signatures",
		EnvVars:  []string{"MCM_TOTAL_SIGNATURES"},
		Required: true,
	}
}
