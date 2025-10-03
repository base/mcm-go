package services

import (
	"context"
	"fmt"

	"mcm-go/pkg/bindings"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/tx"

	"github.com/gagliardetto/solana-go"
)

// SignaturesService handles signature management operations
type SignaturesService struct {
	client *client.Client
}

// NewSignaturesService creates a new signatures service
func NewSignaturesService(client *client.Client) *SignaturesService {
	return &SignaturesService{client: client}
}

// InitSignaturesParams contains parameters for initializing signatures
type InitSignaturesParams struct {
	MultisigID      [32]byte
	Root            [32]byte
	ValidUntil      uint32
	TotalSignatures uint8
}

// InitSignatures initializes the signature storage account
func (s *SignaturesService) InitSignatures(ctx context.Context, params InitSignaturesParams) (solana.Signature, error) {
	ix, err := instructions.InitSignatures(instructions.InitSignaturesParams{
		MultisigID:      params.MultisigID,
		Root:            params.Root,
		ValidUntil:      params.ValidUntil,
		TotalSignatures: params.TotalSignatures,
		Authority:       s.client.Payer.PublicKey(),
		ProgramID:       s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build init signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// AppendSignaturesParams contains parameters for appending signatures
type AppendSignaturesParams struct {
	MultisigID      [32]byte
	Root            [32]byte
	ValidUntil      uint32
	SignaturesBatch []bindings.Signature
}

// AppendSignatures appends a batch of signatures to the storage
func (s *SignaturesService) AppendSignatures(ctx context.Context, params AppendSignaturesParams) (solana.Signature, error) {
	ix, err := instructions.AppendSignatures(instructions.AppendSignaturesParams{
		MultisigID:      params.MultisigID,
		Root:            params.Root,
		ValidUntil:      params.ValidUntil,
		SignaturesBatch: params.SignaturesBatch,
		Authority:       s.client.Payer.PublicKey(),
		ProgramID:       s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build append signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// AppendSignaturesInBatchesParams contains parameters for appending signatures in batches
type AppendSignaturesInBatchesParams struct {
	MultisigID [32]byte
	Root       [32]byte
	ValidUntil uint32
	Signatures []bindings.Signature
	BatchSize  int
}

// AppendSignaturesInBatches appends signatures in multiple batches
func (s *SignaturesService) AppendSignaturesInBatches(
	ctx context.Context,
	params AppendSignaturesInBatchesParams,
) ([]solana.Signature, error) {
	sigs := make([]solana.Signature, 0)

	for i := 0; i < len(params.Signatures); i += params.BatchSize {
		end := i + params.BatchSize
		if end > len(params.Signatures) {
			end = len(params.Signatures)
		}

		batch := params.Signatures[i:end]
		sig, err := s.AppendSignatures(ctx, AppendSignaturesParams{
			MultisigID:      params.MultisigID,
			Root:            params.Root,
			ValidUntil:      params.ValidUntil,
			SignaturesBatch: batch,
		})
		if err != nil {
			return sigs, fmt.Errorf("failed to append signatures batch %d-%d: %w", i, end, err)
		}

		sigs = append(sigs, sig)
	}

	return sigs, nil
}

// FinalizeSignaturesParams contains parameters for finalizing signatures
type FinalizeSignaturesParams struct {
	MultisigID [32]byte
	Root       [32]byte
	ValidUntil uint32
}

// FinalizeSignatures finalizes the signature configuration
func (s *SignaturesService) FinalizeSignatures(ctx context.Context, params FinalizeSignaturesParams) (solana.Signature, error) {
	ix, err := instructions.FinalizeSignatures(instructions.FinalizeSignaturesParams{
		MultisigID: params.MultisigID,
		Root:       params.Root,
		ValidUntil: params.ValidUntil,
		Authority:  s.client.Payer.PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build finalize signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// ClearSignaturesParams contains parameters for clearing signatures
type ClearSignaturesParams struct {
	MultisigID [32]byte
	Root       [32]byte
	ValidUntil uint32
}

// ClearSignatures clears all signatures from the storage
func (s *SignaturesService) ClearSignatures(ctx context.Context, params ClearSignaturesParams) (solana.Signature, error) {
	ix, err := instructions.ClearSignatures(instructions.ClearSignaturesParams{
		MultisigID: params.MultisigID,
		Root:       params.Root,
		ValidUntil: params.ValidUntil,
		Authority:  s.client.Payer.PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build clear signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}
