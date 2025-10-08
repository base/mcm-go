// Package instructions provides high-level wrappers around the generated MCM instruction builders.
package instructions

import (
	"fmt"

	"github.com/base/mcm-go/pkg/bindings"
	"github.com/base/mcm-go/pkg/pda"

	"github.com/gagliardetto/solana-go"
)

// InitializeParams contains parameters for the Initialize instruction
type InitializeParams struct {
	ChainID    uint64
	MultisigID [32]byte
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// Initialize creates an Initialize instruction with all required accounts derived
func Initialize(params InitializeParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	rootMetadata, _, err := pda.RootMetadataPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root metadata PDA: %w", err)
	}

	expiringRootAndOpCount, _, err := pda.ExpiringRootAndOpCountPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive expiring root and op count PDA: %w", err)
	}

	programData, _, err := pda.ProgramDataPDA(params.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive program data PDA: %w", err)
	}

	ix, err := bindings.NewInitializeInstruction(
		params.ChainID,
		params.MultisigID,
		multisigConfig,
		params.Authority,
		solana.SystemProgramID,
		params.ProgramID,
		programData,
		rootMetadata,
		expiringRootAndOpCount,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// TransferOwnershipParams contains parameters for the TransferOwnership instruction
type TransferOwnershipParams struct {
	MultisigID    [32]byte
	ProposedOwner solana.PublicKey
	Authority     solana.PublicKey
	ProgramID     solana.PublicKey
}

// TransferOwnership creates a TransferOwnership instruction with all required accounts derived
func TransferOwnership(params TransferOwnershipParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	ix, err := bindings.NewTransferOwnershipInstruction(
		params.MultisigID,
		params.ProposedOwner,
		multisigConfig,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// AcceptOwnershipParams contains parameters for the AcceptOwnership instruction
type AcceptOwnershipParams struct {
	MultisigID [32]byte
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// AcceptOwnership creates an AcceptOwnership instruction with all required accounts derived
func AcceptOwnership(params AcceptOwnershipParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	ix, err := bindings.NewAcceptOwnershipInstruction(
		params.MultisigID,
		multisigConfig,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// SetConfigParams contains parameters for the SetConfig instruction
type SetConfigParams struct {
	MultisigID   [32]byte
	SignerGroups []byte
	GroupQuorums [32]uint8
	GroupParents [32]uint8
	ClearRoot    bool
	Authority    solana.PublicKey
	ProgramID    solana.PublicKey
}

// SetConfig creates a SetConfig instruction
func SetConfig(params SetConfigParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	configSigners, _, err := pda.MultisigConfigSignersPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	rootMetadata, _, err := pda.RootMetadataPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root metadata PDA: %w", err)
	}

	expiringRootAndOpCount, _, err := pda.ExpiringRootAndOpCountPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive expiring root and op count PDA: %w", err)
	}

	ix, err := bindings.NewSetConfigInstruction(
		params.MultisigID,
		params.SignerGroups,
		params.GroupQuorums,
		params.GroupParents,
		params.ClearRoot,
		multisigConfig,
		configSigners,
		rootMetadata,
		expiringRootAndOpCount,
		params.Authority,
		solana.SystemProgramID,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// SetRootParams contains parameters for the SetRoot instruction
type SetRootParams struct {
	MultisigID    [32]byte
	Root          [32]byte
	ValidUntil    uint32
	Metadata      bindings.RootMetadataInput
	MetadataProof [][32]uint8
	Authority     solana.PublicKey
	ProgramID     solana.PublicKey
}

// SetRoot creates a SetRoot instruction
func SetRoot(params SetRootParams) (solana.Instruction, error) {
	rootSignatures, _, err := pda.RootSignaturesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Authority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root signatures PDA: %w", err)
	}

	rootMetadata, _, err := pda.RootMetadataPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root metadata PDA: %w", err)
	}

	seenSignedHashes, _, err := pda.SeenSignedHashesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive seen signed hashes PDA: %w", err)
	}

	expiringRootAndOpCount, _, err := pda.ExpiringRootAndOpCountPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive expiring root and op count PDA: %w", err)
	}

	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	ix, err := bindings.NewSetRootInstruction(
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Metadata,
		params.MetadataProof,
		rootSignatures,
		rootMetadata,
		seenSignedHashes,
		expiringRootAndOpCount,
		multisigConfig,
		params.Authority,
		solana.SystemProgramID,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// ExecuteParams contains parameters for the Execute instruction
type ExecuteParams struct {
	MultisigID        [32]byte
	ChainID           uint64
	Nonce             uint64
	To                solana.PublicKey
	Data              []byte
	Proof             [][32]uint8
	RemainingAccounts []*solana.AccountMeta // Accounts that will be passed to the target program
	Authority         solana.PublicKey
	ProgramID         solana.PublicKey
}

// Execute creates an Execute instruction
func Execute(params ExecuteParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	rootMetadata, _, err := pda.RootMetadataPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root metadata PDA: %w", err)
	}

	expiringRootAndOpCount, _, err := pda.ExpiringRootAndOpCountPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive expiring root and op count PDA: %w", err)
	}

	multisigSigner, _, err := pda.MultisigSignerPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig signer PDA: %w", err)
	}

	ix, err := bindings.NewExecuteInstruction(
		params.MultisigID,
		params.ChainID,
		params.Nonce,
		params.Data,
		params.Proof,
		multisigConfig,
		rootMetadata,
		expiringRootAndOpCount,
		params.To,
		multisigSigner,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	// Append remaining accounts that will be passed to the target program via CPI
	// These accounts are part of the Merkle leaf hash and must match exactly
	if len(params.RemainingAccounts) > 0 {
		ixImpl := ix.(*solana.GenericInstruction)
		ixImpl.AccountValues = append(ixImpl.AccountValues, params.RemainingAccounts...)
	}

	return setProgramID(ix, params.ProgramID), nil
}

// InitSignersParams contains parameters for the InitSigners instruction
type InitSignersParams struct {
	MultisigID   [32]byte
	TotalSigners uint8
	Authority    solana.PublicKey
	ProgramID    solana.PublicKey
}

// InitSigners creates an InitSigners instruction
func InitSigners(params InitSignersParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	configSigners, _, err := pda.MultisigConfigSignersPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	ix, err := bindings.NewInitSignersInstruction(
		params.MultisigID,
		params.TotalSigners,
		multisigConfig,
		configSigners,
		params.Authority,
		solana.SystemProgramID,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// AppendSignersParams contains parameters for the AppendSigners instruction
type AppendSignersParams struct {
	MultisigID   [32]byte
	SignersBatch [][20]uint8
	Authority    solana.PublicKey
	ProgramID    solana.PublicKey
}

// AppendSigners creates an AppendSigners instruction
func AppendSigners(params AppendSignersParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	configSigners, _, err := pda.MultisigConfigSignersPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	ix, err := bindings.NewAppendSignersInstruction(
		params.MultisigID,
		params.SignersBatch,
		multisigConfig,
		configSigners,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// FinalizeSignersParams contains parameters for the FinalizeSigners instruction
type FinalizeSignersParams struct {
	MultisigID [32]byte
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// FinalizeSigners creates a FinalizeSigners instruction
func FinalizeSigners(params FinalizeSignersParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	configSigners, _, err := pda.MultisigConfigSignersPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	ix, err := bindings.NewFinalizeSignersInstruction(
		params.MultisigID,
		multisigConfig,
		configSigners,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// InitSignaturesParams contains parameters for the InitSignatures instruction
type InitSignaturesParams struct {
	MultisigID      [32]byte
	Root            [32]byte
	ValidUntil      uint32
	TotalSignatures uint8
	Authority       solana.PublicKey
	ProgramID       solana.PublicKey
}

// InitSignatures creates an InitSignatures instruction
func InitSignatures(params InitSignaturesParams) (solana.Instruction, error) {
	signatures, _, err := pda.RootSignaturesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Authority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive signatures PDA: %w", err)
	}

	ix, err := bindings.NewInitSignaturesInstruction(
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.TotalSignatures,
		signatures,
		params.Authority,
		solana.SystemProgramID,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// AppendSignaturesParams contains parameters for the AppendSignatures instruction
type AppendSignaturesParams struct {
	MultisigID      [32]byte
	Root            [32]byte
	ValidUntil      uint32
	SignaturesBatch []bindings.Signature
	Authority       solana.PublicKey
	ProgramID       solana.PublicKey
}

// AppendSignatures creates an AppendSignatures instruction
func AppendSignatures(params AppendSignaturesParams) (solana.Instruction, error) {
	signatures, _, err := pda.RootSignaturesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Authority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive signatures PDA: %w", err)
	}

	ix, err := bindings.NewAppendSignaturesInstruction(
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.SignaturesBatch,
		signatures,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// FinalizeSignaturesParams contains parameters for the FinalizeSignatures instruction
type FinalizeSignaturesParams struct {
	MultisigID [32]byte
	Root       [32]byte
	ValidUntil uint32
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// FinalizeSignatures creates a FinalizeSignatures instruction
func FinalizeSignatures(params FinalizeSignaturesParams) (solana.Instruction, error) {
	signatures, _, err := pda.RootSignaturesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Authority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive signatures PDA: %w", err)
	}

	ix, err := bindings.NewFinalizeSignaturesInstruction(
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		signatures,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// ClearSignaturesParams contains parameters for the ClearSignatures instruction
type ClearSignaturesParams struct {
	MultisigID [32]byte
	Root       [32]byte
	ValidUntil uint32
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// ClearSignatures creates a ClearSignatures instruction
func ClearSignatures(params ClearSignaturesParams) (solana.Instruction, error) {
	signatures, _, err := pda.RootSignaturesPDA(
		params.ProgramID,
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		params.Authority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive signatures PDA: %w", err)
	}

	ix, err := bindings.NewClearSignaturesInstruction(
		params.MultisigID,
		params.Root,
		params.ValidUntil,
		signatures,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// ClearSignersParams contains parameters for the ClearSigners instruction
type ClearSignersParams struct {
	MultisigID [32]byte
	Authority  solana.PublicKey
	ProgramID  solana.PublicKey
}

// ClearSigners creates a ClearSigners instruction
func ClearSigners(params ClearSignersParams) (solana.Instruction, error) {
	multisigConfig, _, err := pda.MultisigConfigPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig config PDA: %w", err)
	}

	configSigners, _, err := pda.MultisigConfigSignersPDA(params.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	ix, err := bindings.NewClearSignersInstruction(
		params.MultisigID,
		multisigConfig,
		configSigners,
		params.Authority,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// setProgramID overrides the ProgramID in the instruction to use the user-configured value
// instead of the hardcoded value from bindings.
func setProgramID(ix solana.Instruction, programID solana.PublicKey) solana.Instruction {
	if genericIx, ok := ix.(*solana.GenericInstruction); ok {
		genericIx.ProgID = programID
	}
	return ix
}
