// Package state provides utilities for fetching and managing on-chain account state.
package state

import (
	"context"
	"fmt"

	"github.com/base/mcm-go/pkg/bindings"
	"github.com/base/mcm-go/pkg/pda"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Fetcher provides methods to fetch MCM program accounts
type Fetcher struct {
	rpc       *rpc.Client
	programID solana.PublicKey
}

// NewFetcher creates a new state fetcher
func NewFetcher(rpcClient *rpc.Client, programID solana.PublicKey) *Fetcher {
	return &Fetcher{
		rpc:       rpcClient,
		programID: programID,
	}
}

// GetMultisigConfig fetches the MultisigConfig account
func (f *Fetcher) GetMultisigConfig(ctx context.Context, multisigID [32]byte) (*bindings.MultisigConfig, error) {
	configPDA, _, err := pda.MultisigConfigPDA(f.programID, multisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, configPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	config, err := bindings.ParseAccount_MultisigConfig(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse MultisigConfig: %w", err)
	}

	return config, nil
}

// GetConfigSigners fetches the ConfigSigners account
func (f *Fetcher) GetConfigSigners(ctx context.Context, multisigID [32]byte) (*bindings.ConfigSigners, error) {
	signersPDA, _, err := pda.MultisigConfigSignersPDA(f.programID, multisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config signers PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, signersPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	signers, err := bindings.ParseAccount_ConfigSigners(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ConfigSigners: %w", err)
	}

	return signers, nil
}

// GetRootMetadata fetches the RootMetadata account
func (f *Fetcher) GetRootMetadata(ctx context.Context, multisigID [32]byte) (*bindings.RootMetadata, error) {
	metadataPDA, _, err := pda.RootMetadataPDA(f.programID, multisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root metadata PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, metadataPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	metadata, err := bindings.ParseAccount_RootMetadata(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse RootMetadata: %w", err)
	}

	return metadata, nil
}

// GetExpiringRootAndOpCount fetches the ExpiringRootAndOpCount account
func (f *Fetcher) GetExpiringRootAndOpCount(ctx context.Context, multisigID [32]byte) (*bindings.ExpiringRootAndOpCount, error) {
	rootPDA, _, err := pda.ExpiringRootAndOpCountPDA(f.programID, multisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive expiring root PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, rootPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	root, err := bindings.ParseAccount_ExpiringRootAndOpCount(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ExpiringRootAndOpCount: %w", err)
	}

	return root, nil
}

// GetRootSignatures fetches the RootSignatures account
func (f *Fetcher) GetRootSignatures(
	ctx context.Context,
	multisigID [32]byte,
	root [32]byte,
	validUntil uint32,
	authority solana.PublicKey,
) (*bindings.RootSignatures, error) {
	signaturesPDA, _, err := pda.RootSignaturesPDA(f.programID, multisigID, root, validUntil, authority)
	if err != nil {
		return nil, fmt.Errorf("failed to derive root signatures PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, signaturesPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	signatures, err := bindings.ParseAccount_RootSignatures(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse RootSignatures: %w", err)
	}

	return signatures, nil
}

// GetSeenSignedHash fetches the SeenSignedHash account
func (f *Fetcher) GetSeenSignedHash(
	ctx context.Context,
	multisigID [32]byte,
	root [32]byte,
	validUntil uint32,
) (*bindings.SeenSignedHash, error) {
	seenPDA, _, err := pda.SeenSignedHashesPDA(f.programID, multisigID, root, validUntil)
	if err != nil {
		return nil, fmt.Errorf("failed to derive seen signed hash PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, seenPDA)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("account not found")
	}

	seen, err := bindings.ParseAccount_SeenSignedHash(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse SeenSignedHash: %w", err)
	}

	return seen, nil
}
