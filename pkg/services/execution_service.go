package services

import (
	"context"
	"fmt"

	"mcm-go/pkg/client"
	"mcm-go/pkg/instructions"
	"mcm-go/pkg/proposal"
	"mcm-go/pkg/tx"

	"github.com/gagliardetto/solana-go"
)

// ExecutionService handles operation execution
type ExecutionService struct {
	client *client.Client
}

// NewExecutionService creates a new execution service
func NewExecutionService(client *client.Client) *ExecutionService {
	return &ExecutionService{client: client}
}

// ExecuteOperationParams contains parameters for executing an operation
type ExecuteOperationParams struct {
	MultisigID       [32]byte
	ProposalWithRoot *proposal.ProposalWithRoot
	OperationIndex   int
}

// ExecuteOperation executes a single operation from a proposal
func (s *ExecutionService) ExecuteOperation(ctx context.Context, params ExecuteOperationParams) (solana.Signature, error) {
	if params.OperationIndex >= len(params.ProposalWithRoot.Instructions) {
		return solana.Signature{}, fmt.Errorf("operation index %d out of range", params.OperationIndex)
	}

	if params.OperationIndex >= len(params.ProposalWithRoot.OperationProofs) {
		return solana.Signature{}, fmt.Errorf("operation proof index %d out of range", params.OperationIndex)
	}

	op := params.ProposalWithRoot.Instructions[params.OperationIndex]
	proof := params.ProposalWithRoot.OperationProofs[params.OperationIndex]

	// Calculate nonce (preOpCount + index)
	nonce := params.ProposalWithRoot.RootMetadata.PreOpCount + uint64(params.OperationIndex)

	// Convert proof to [][32]uint8
	proofArray := make([][32]uint8, len(proof))
	copy(proofArray, proof)

	ix, err := instructions.Execute(instructions.ExecuteParams{
		MultisigID:        params.MultisigID,
		ChainID:           params.ProposalWithRoot.RootMetadata.ChainID,
		Nonce:             nonce,
		To:                op.ProgID,
		Data:              op.DataBytes,
		Proof:             proofArray,
		RemainingAccounts: op.AccountValues,
		Authority:         s.client.Payer.PublicKey(),
		ProgramID:         s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build execute instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS, s.client.Payer).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// ExecuteAllOperationsParams contains parameters for executing all operations
type ExecuteAllOperationsParams struct {
	MultisigID       [32]byte
	ProposalWithRoot *proposal.ProposalWithRoot
}

// ExecuteAllOperations executes all operations in a proposal sequentially
func (s *ExecutionService) ExecuteAllOperations(ctx context.Context, params ExecuteAllOperationsParams) ([]solana.Signature, error) {
	sigs := make([]solana.Signature, 0, len(params.ProposalWithRoot.Instructions))

	for i := range params.ProposalWithRoot.Instructions {
		sig, err := s.ExecuteOperation(ctx, ExecuteOperationParams{
			MultisigID:       params.MultisigID,
			ProposalWithRoot: params.ProposalWithRoot,
			OperationIndex:   i,
		})
		if err != nil {
			return sigs, fmt.Errorf("failed to execute operation %d: %w", i, err)
		}

		sigs = append(sigs, sig)
	}

	return sigs, nil
}
