// Package pda provides utilities for deriving Program Derived Addresses (PDAs)
// for the MCM (Multi-Chain Multisig) program on Solana.
package pda

import (
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

const (
	// BPFLoaderUpgradeableProgramID is the program ID for the BPF upgradeable loader
	BPFLoaderUpgradeableProgramID = "BPFLoaderUpgradeab1e11111111111111111111111"
)

var (
	// Seed constants
	signerSeed                 = []byte("multisig_signer")
	configSeed                 = []byte("multisig_config")
	configSignersSeed          = []byte("multisig_config_signers")
	rootMetadataSeed           = []byte("root_metadata")
	rootSignaturesSeed         = []byte("root_signatures")
	expiringRootAndOpCountSeed = []byte("expiring_root_and_op_count")
	seenSignedHashesSeed       = []byte("seen_signed_hashes")
)

// MultisigSignerPDA derives the PDA for the multisig signer
func MultisigSignerPDA(mcmProgram solana.PublicKey, multisigID [32]byte) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			signerSeed,
			multisigID[:],
		},
		mcmProgram,
	)
}

// MultisigConfigPDA derives the PDA for the multisig configuration
func MultisigConfigPDA(mcmProgram solana.PublicKey, multisigID [32]byte) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			configSeed,
			multisigID[:],
		},
		mcmProgram,
	)
}

// MultisigConfigSignersPDA derives the PDA for the multisig config signers
func MultisigConfigSignersPDA(mcmProgram solana.PublicKey, multisigID [32]byte) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			configSignersSeed,
			multisigID[:],
		},
		mcmProgram,
	)
}

// RootMetadataPDA derives the PDA for root metadata
func RootMetadataPDA(mcmProgram solana.PublicKey, multisigID [32]byte) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			rootMetadataSeed,
			multisigID[:],
		},
		mcmProgram,
	)
}

// ExpiringRootAndOpCountPDA derives the PDA for expiring root and operation count
func ExpiringRootAndOpCountPDA(mcmProgram solana.PublicKey, multisigID [32]byte) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			expiringRootAndOpCountSeed,
			multisigID[:],
		},
		mcmProgram,
	)
}

// RootSignaturesPDA derives the PDA for root signatures
func RootSignaturesPDA(
	mcmProgram solana.PublicKey,
	multisigID [32]byte,
	root [32]byte,
	validUntil uint32,
	authority solana.PublicKey,
) (solana.PublicKey, uint8, error) {
	validUntilBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(validUntilBytes, validUntil)

	return solana.FindProgramAddress(
		[][]byte{
			rootSignaturesSeed,
			multisigID[:],
			root[:],
			validUntilBytes,
			authority[:],
		},
		mcmProgram,
	)
}

// SeenSignedHashesPDA derives the PDA for seen signed hashes
func SeenSignedHashesPDA(
	mcmProgram solana.PublicKey,
	multisigID [32]byte,
	root [32]byte,
	validUntil uint32,
) (solana.PublicKey, uint8, error) {
	validUntilBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(validUntilBytes, validUntil)

	return solana.FindProgramAddress(
		[][]byte{
			seenSignedHashesSeed,
			multisigID[:],
			root[:],
			validUntilBytes,
		},
		mcmProgram,
	)
}

// ProgramDataPDA derives the PDA for program data account
func ProgramDataPDA(programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	loaderProgramID, err := solana.PublicKeyFromBase58(BPFLoaderUpgradeableProgramID)
	if err != nil {
		return solana.PublicKey{}, 0, fmt.Errorf("invalid BPF loader program ID: %w", err)
	}

	return solana.FindProgramAddress(
		[][]byte{
			programID[:],
		},
		loaderProgramID,
	)
}
