// Package state provides utilities for fetching and managing on-chain account state for the Bridge program.
package state

import (
	"context"
	"fmt"

	"github.com/base/mcm-go/cmd/internal/bridge/bindings"
	bridgePDA "github.com/base/mcm-go/cmd/internal/bridge/pda"
	mcmPDA "github.com/base/mcm-go/pkg/pda"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Fetcher provides methods to fetch Bridge program accounts
type Fetcher struct {
	rpc       *rpc.Client
	programID solana.PublicKey
}

// NewFetcher creates a new state fetcher for the Bridge program
func NewFetcher(rpcClient *rpc.Client, programID solana.PublicKey) *Fetcher {
	return &Fetcher{
		rpc:       rpcClient,
		programID: programID,
	}
}

// GetBridge fetches the Bridge account
func (f *Fetcher) GetBridge(ctx context.Context) (*bindings.Bridge, error) {
	bridgeAddr, _, err := bridgePDA.BridgePDA(f.programID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bridge PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, bridgeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("bridge account not found")
	}

	bridge, err := bindings.ParseAccount_Bridge(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Bridge account: %w", err)
	}

	return bridge, nil
}

// GetUpgradeAuthority fetches the upgrade authority from the bridge program's ProgramData account
// TODO: This could be factorized in some internal/loader_v3 package
func (f *Fetcher) GetUpgradeAuthority(ctx context.Context) (solana.PublicKey, error) {
	programData, _, err := mcmPDA.ProgramDataPDA(f.programID)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to derive program data PDA: %w", err)
	}

	accountInfo, err := f.rpc.GetAccountInfo(ctx, programData)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to get program data account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("program data account not found")
	}

	// ProgramData account structure (BPF Loader Upgradeable):
	// Serialized with bincode as an enum:
	// - 4 bytes: enum discriminator (3 = ProgramData)
	// - 8 bytes: slot (u64)
	// - 1 byte: option flag for upgrade authority (0 = None, 1 = Some)
	// - 32 bytes: upgrade authority pubkey (if option = 1)
	data := accountInfo.Value.Data.GetBinary()
	if len(data) < 45 {
		return solana.PublicKey{}, fmt.Errorf("invalid program data account size")
	}

	// Check if upgrade authority is set (byte 12 should be 1)
	if data[12] != 1 {
		return solana.PublicKey{}, fmt.Errorf("program has no upgrade authority (immutable)")
	}

	// Extract upgrade authority (bytes 13-45)
	var upgradeAuthority solana.PublicKey
	copy(upgradeAuthority[:], data[13:45])

	return upgradeAuthority, nil
}
