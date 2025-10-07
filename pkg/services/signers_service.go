package services

import (
	"context"
	"fmt"

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
	ix, err := instructions.AppendSigners(instructions.AppendSignersParams{
		MultisigID:   params.MultisigID,
		SignersBatch: params.SignersBatch,
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
