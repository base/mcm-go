// Package instructions provides high-level wrappers around the generated Bridge instruction builders.
package instructions

import (
	"fmt"

	"github.com/base/mcm-go/cmd/internal/bridge/bindings"
	"github.com/base/mcm-go/cmd/internal/bridge/pda"
	mcmPDA "github.com/base/mcm-go/pkg/pda"

	"github.com/gagliardetto/solana-go"
)

// SetPauseStatusParams contains parameters for the SetPauseStatus instruction
type SetPauseStatusParams struct {
	Paused    bool
	Guardian  solana.PublicKey
	ProgramID solana.PublicKey
}

// SetPauseStatus creates a SetPauseStatus instruction with all required accounts derived
func SetPauseStatus(params SetPauseStatusParams) (solana.Instruction, error) {
	bridgeAddr, _, err := pda.BridgePDA(params.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bridge PDA: %w", err)
	}

	ix, err := bindings.NewSetPauseStatusInstruction(
		params.Paused,
		bridgeAddr,
		params.Guardian,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// SetPartnerOracleConfigParams contains parameters for the SetPartnerOracleConfig instruction
type SetPartnerOracleConfigParams struct {
	RequiredThreshold uint8
	UpgradeAuthority  solana.PublicKey
	ProgramID         solana.PublicKey
}

// SetPartnerOracleConfig creates a SetPartnerOracleConfig instruction with all required accounts derived
func SetPartnerOracleConfig(params SetPartnerOracleConfigParams) (solana.Instruction, error) {
	bridgeAddr, _, err := pda.BridgePDA(params.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bridge PDA: %w", err)
	}

	// Derive program data PDA (needed for upgrade authority verification)
	programData, _, err := mcmPDA.ProgramDataPDA(params.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive program data PDA: %w", err)
	}

	ix, err := bindings.NewSetPartnerOracleConfigInstruction(
		bindings.PartnerOracleConfig{
			RequiredThreshold: params.RequiredThreshold,
		},
		params.UpgradeAuthority,
		bridgeAddr,
		programData,
		params.ProgramID,
	)
	if err != nil {
		return nil, err
	}

	return setProgramID(ix, params.ProgramID), nil
}

// setProgramID overrides the program ID in the instruction.
// This is critical because the program ID in the IDL may be for dev/test,
// but we need to use the actual program ID provided by the user.
func setProgramID(ix solana.Instruction, programID solana.PublicKey) solana.Instruction {
	if genericIx, ok := ix.(*solana.GenericInstruction); ok {
		genericIx.ProgID = programID
	}
	return ix
}
