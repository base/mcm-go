// Package services provides high-level services for managing MCM operations.
package services

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/tx"
)

// SignersService handles signer configuration operations
type SignersService struct {
	client *client.Client
}

// NewSignersService creates a new signers service
func NewSignersService(client *client.Client) *SignersService {
	return &SignersService{client: client}
}

// InitSignersParams contains parameters for initializing signers
type InitSignersParams struct {
	MultisigID   [32]byte
	TotalSigners uint8
	Authority    solana.PublicKey
}

// InitSigners initializes the signer storage account
func (s *SignersService) InitSigners(ctx context.Context, params InitSignersParams) (solana.Signature, error) {
	ix, err := instructions.InitSigners(instructions.InitSignersParams{
		MultisigID:   params.MultisigID,
		TotalSigners: params.TotalSigners,
		Authority:    params.Authority,
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build init signers instruction: %w", err)
	}

	builder := tx.New(s.client.RPC, s.client.WS, s.client.GetPayer())
	builder.AddInstruction(ix)

	if s.client.DefaultPayer != nil {
		builder.AddSigner(*s.client.DefaultPayer)
	}

	return builder.BuildSignAndSendWithConfirmation(ctx)
}

// AppendSignersParams contains parameters for appending signers
type AppendSignersParams struct {
	MultisigID   [32]byte
	SignersBatch [][20]uint8
	Authority    solana.PublicKey
}

// AppendSigners appends a batch of signers to the storage
func (s *SignersService) AppendSigners(ctx context.Context, params AppendSignersParams) (solana.Signature, error) {
	ix, err := instructions.AppendSigners(instructions.AppendSignersParams{
		MultisigID:   params.MultisigID,
		SignersBatch: params.SignersBatch,
		Authority:    params.Authority,
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build append signers instruction: %w", err)
	}

	builder := tx.New(s.client.RPC, s.client.WS, s.client.GetPayer())
	builder.AddInstruction(ix)

	if s.client.DefaultPayer != nil {
		builder.AddSigner(*s.client.DefaultPayer)
	}

	return builder.BuildSignAndSendWithConfirmation(ctx)
}

// AppendSignersInBatchesParams contains parameters for appending signers in batches
type AppendSignersInBatchesParams struct {
	MultisigID [32]byte
	Signers    [][20]uint8
	BatchSize  int
	Authority  solana.PublicKey
}

// AppendSignersInBatches appends signers in multiple batches to handle large signer sets
func (s *SignersService) AppendSignersInBatches(
	ctx context.Context,
	params AppendSignersInBatchesParams,
) ([]solana.Signature, error) {
	sigs := make([]solana.Signature, 0)

	for i := 0; i < len(params.Signers); i += params.BatchSize {
		end := i + params.BatchSize
		if end > len(params.Signers) {
			end = len(params.Signers)
		}

		batch := params.Signers[i:end]
		sig, err := s.AppendSigners(ctx, AppendSignersParams{
			MultisigID:   params.MultisigID,
			SignersBatch: batch,
			Authority:    params.Authority,
		})
		if err != nil {
			return sigs, fmt.Errorf("failed to append batch %d-%d: %w", i, end, err)
		}

		sigs = append(sigs, sig)
	}

	return sigs, nil
}

// FinalizeSignersParams contains parameters for finalizing signers
type FinalizeSignersParams struct {
	MultisigID [32]byte
	Authority  solana.PublicKey
}

// FinalizeSigners finalizes the signer configuration
func (s *SignersService) FinalizeSigners(ctx context.Context, params FinalizeSignersParams) (solana.Signature, error) {
	ix, err := instructions.FinalizeSigners(instructions.FinalizeSignersParams{
		MultisigID: params.MultisigID,
		Authority:  params.Authority,
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build finalize signers instruction: %w", err)
	}

	builder := tx.New(s.client.RPC, s.client.WS, s.client.GetPayer())
	builder.AddInstruction(ix)

	if s.client.DefaultPayer != nil {
		builder.AddSigner(*s.client.DefaultPayer)
	}

	return builder.BuildSignAndSendWithConfirmation(ctx)
}

// SetConfigParams contains parameters for setting config
type SetConfigParams struct {
	MultisigID   [32]byte
	SignerGroups []byte
	GroupQuorums [32]uint8
	GroupParents [32]uint8
	ClearRoot    bool
	Authority    solana.PublicKey
}

// SetConfig sets the multisig configuration with signer groups and quorums
func (s *SignersService) SetConfig(ctx context.Context, params SetConfigParams) (solana.Signature, error) {
	ix, err := instructions.SetConfig(instructions.SetConfigParams{
		MultisigID:   params.MultisigID,
		SignerGroups: params.SignerGroups,
		GroupQuorums: params.GroupQuorums,
		GroupParents: params.GroupParents,
		ClearRoot:    params.ClearRoot,
		Authority:    params.Authority,
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build set config instruction: %w", err)
	}

	builder := tx.New(s.client.RPC, s.client.WS, s.client.GetPayer())
	builder.AddInstruction(ix)

	if s.client.DefaultPayer != nil {
		builder.AddSigner(*s.client.DefaultPayer)
	}

	return builder.BuildSignAndSendWithConfirmation(ctx)
}

// ClearSignersParams contains parameters for clearing signers
type ClearSignersParams struct {
	MultisigID [32]byte
	Authority  solana.PublicKey
}

// ClearSigners clears all signers from the configuration
func (s *SignersService) ClearSigners(ctx context.Context, params ClearSignersParams) (solana.Signature, error) {
	ix, err := instructions.ClearSigners(instructions.ClearSignersParams{
		MultisigID: params.MultisigID,
		Authority:  params.Authority,
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build clear signers instruction: %w", err)
	}

	builder := tx.New(s.client.RPC, s.client.WS, s.client.GetPayer())
	builder.AddInstruction(ix)

	if s.client.DefaultPayer != nil {
		builder.AddSigner(*s.client.DefaultPayer)
	}

	return builder.BuildSignAndSendWithConfirmation(ctx)
}
