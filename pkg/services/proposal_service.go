package services

import (
	"context"
	"fmt"

	"mcm-go/pkg/bindings"
	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/pda"
	"mcm-go/pkg/proposal"
	"mcm-go/pkg/state"
	"mcm-go/pkg/tx"

	"github.com/gagliardetto/solana-go"
)

// ProposalService handles proposal creation and root setting operations
type ProposalService struct {
	client *client.Client
}

// NewProposalService creates a new proposal service
func NewProposalService(client *client.Client) *ProposalService {
	return &ProposalService{client: client}
}

// SetRootParams contains parameters for setting a new root
type SetRootParams struct {
	MultisigID [32]byte
	Proposal   *proposal.ProposalWithRoot
}

// SetRoot sets a new Merkle root for the multisig
func (s *ProposalService) SetRoot(ctx context.Context, params SetRootParams) (solana.Signature, error) {
	// Convert ProposalWithRoot metadata to RootMetadataInput
	metadata := bindings.RootMetadataInput{
		ChainId:              params.Proposal.RootMetadata.ChainID,
		Multisig:             params.Proposal.RootMetadata.Multisig,
		PreOpCount:           params.Proposal.RootMetadata.PreOpCount,
		PostOpCount:          params.Proposal.RootMetadata.PostOpCount,
		OverridePreviousRoot: params.Proposal.RootMetadata.OverridePreviousRoot,
	}

	// Convert metadata proof to [][32]uint8
	metadataProof := make([][32]uint8, len(params.Proposal.MetadataProof))
	copy(metadataProof, params.Proposal.MetadataProof)

	ix, err := instructions.SetRoot(instructions.SetRootParams{
		MultisigID:    params.MultisigID,
		Root:          params.Proposal.Root,
		ValidUntil:    params.Proposal.ValidUntil,
		Metadata:      metadata,
		MetadataProof: metadataProof,
		Authority:     s.client.Payer.PublicKey(),
		ProgramID:     s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build set root instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// CreateProposalFromChainParams contains parameters for creating a proposal from on-chain state
type CreateProposalFromChainParams struct {
	MultisigID           [32]byte
	ValidUntil           uint32
	Instructions         []*solana.GenericInstruction
	OverridePreviousRoot bool
}

// CreateProposalFromChain creates a proposal by fetching the current on-chain state
// and deriving the metadata automatically.
//
// This method:
// 1. Fetches MultisigConfig for chain_id
// 2. Fetches ExpiringRootAndOpCount for the current op_count
// 3. Derives the MultisigConfig PDA as the multisig address
// 4. Calculates pre_op_count (current) and post_op_count (current + num_instructions)
// 5. Returns the proposal - caller can use WithRoot() and WithHashToSign() as needed
//
// The caller is responsible for ensuring valid_until is appropriate and
// deciding whether to override the previous root.
func (s *ProposalService) CreateProposalFromChain(
	ctx context.Context,
	params CreateProposalFromChainParams,
) (*proposal.Proposal, error) {
	fetcher := state.NewFetcher(s.client.RPC, s.client.ProgramID)

	// Fetch only what we need: config and current op count
	config, err := fetcher.GetMultisigConfig(ctx, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch multisig config: %w", err)
	}

	expiringRoot, err := fetcher.GetExpiringRootAndOpCount(ctx, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expiring root and op count: %w", err)
	}

	// Derive the MultisigConfig PDA - this is the multisig address
	configPDA, _, err := pda.MultisigConfigPDA(s.client.ProgramID, params.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive config PDA: %w", err)
	}

	// Build metadata from on-chain state
	preOpCount := expiringRoot.OpCount
	postOpCount := preOpCount + uint64(len(params.Instructions))

	metadata := proposal.RootMetadata{
		ChainID:              config.ChainId,
		Multisig:             configPDA,
		PreOpCount:           preOpCount,
		PostOpCount:          postOpCount,
		OverridePreviousRoot: params.OverridePreviousRoot,
	}

	// Delegate to existing builder
	builder := proposal.NewBuilder(params.MultisigID, params.ValidUntil)
	builder.SetRootMetadata(metadata)

	for _, ix := range params.Instructions {
		builder.AddInstruction(ix)
	}

	return builder.Build()
}
