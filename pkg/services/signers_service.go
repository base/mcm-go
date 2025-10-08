package services

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/base/mcm-go/pkg/client"
	"github.com/base/mcm-go/pkg/instructions"

	"github.com/gagliardetto/solana-go"
)

type SignersService struct {
	client *client.Client
}

func NewSignersService(client *client.Client) *SignersService {
	return &SignersService{client: client}
}

type InitSignersParams struct {
	MultisigID   [32]byte
	TotalSigners uint8
}

func (s *SignersService) InitSigners(ctx context.Context, params InitSignersParams) (solana.Signature, error) {
	ix, err := instructions.InitSigners(instructions.InitSignersParams{
		MultisigID:   params.MultisigID,
		TotalSigners: params.TotalSigners,
		Authority:    s.client.Payer().PublicKey(),
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build init signers instruction: %w", err)
	}

	return s.client.BuildSignAndSendWithConfirmation(ctx, ix)
}

type AppendSignersParams struct {
	MultisigID   [32]byte
	SignersBatch [][20]uint8
}

func (s *SignersService) AppendSigners(ctx context.Context, params AppendSignersParams) (solana.Signature, error) {
	// Sort signers by address (strictly increasing order) as required by the Solana contract
	sortedSigners, err := sortSigners(params.SignersBatch)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sort signers: %w", err)
	}

	ix, err := instructions.AppendSigners(instructions.AppendSignersParams{
		MultisigID:   params.MultisigID,
		SignersBatch: sortedSigners,
		Authority:    s.client.Payer().PublicKey(),
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build append signers instruction: %w", err)
	}

	return s.client.BuildSignAndSendWithConfirmation(ctx, ix)
}

type FinalizeSignersParams struct {
	MultisigID [32]byte
}

func (s *SignersService) FinalizeSigners(ctx context.Context, params FinalizeSignersParams) (solana.Signature, error) {
	ix, err := instructions.FinalizeSigners(instructions.FinalizeSignersParams{
		MultisigID: params.MultisigID,
		Authority:  s.client.Payer().PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build finalize signers instruction: %w", err)
	}

	return s.client.BuildSignAndSendWithConfirmation(ctx, ix)
}

type ClearSignersParams struct {
	MultisigID [32]byte
}

func (s *SignersService) ClearSigners(ctx context.Context, params ClearSignersParams) (solana.Signature, error) {
	ix, err := instructions.ClearSigners(instructions.ClearSignersParams{
		MultisigID: params.MultisigID,
		Authority:  s.client.Payer().PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build clear signers instruction: %w", err)
	}

	return s.client.BuildSignAndSendWithConfirmation(ctx, ix)
}

type SetConfigParams struct {
	MultisigID   [32]byte
	SignerGroups []byte
	GroupQuorums [32]uint8
	GroupParents [32]uint8
	ClearRoot    bool
}

func (s *SignersService) SetConfig(ctx context.Context, params SetConfigParams) (solana.Signature, error) {
	ix, err := instructions.SetConfig(instructions.SetConfigParams{
		MultisigID:   params.MultisigID,
		SignerGroups: params.SignerGroups,
		GroupQuorums: params.GroupQuorums,
		GroupParents: params.GroupParents,
		ClearRoot:    params.ClearRoot,
		Authority:    s.client.Payer().PublicKey(),
		ProgramID:    s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build set config instruction: %w", err)
	}

	return s.client.BuildSignAndSendWithConfirmation(ctx, ix)
}

// sortSigners sorts a batch of signers by address in strictly increasing order.
// Returns an error if duplicate addresses are detected.
func sortSigners(signers [][20]uint8) ([][20]uint8, error) {
	if len(signers) == 0 {
		return signers, nil
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([][20]uint8, len(signers))
	copy(sorted, signers)

	// Sort by address in ascending order
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) < 0
	})

	// Validate strictly increasing order (no duplicates)
	for i := 1; i < len(sorted); i++ {
		if bytes.Equal(sorted[i-1][:], sorted[i][:]) {
			return nil, fmt.Errorf("duplicate signer address detected: %x", sorted[i])
		}
	}

	return sorted, nil
}
