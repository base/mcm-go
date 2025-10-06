package services

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/base/mcm-go/pkg/bindings"
	"github.com/base/mcm-go/pkg/client"
	"github.com/base/mcm-go/pkg/crypto"
	"github.com/base/mcm-go/pkg/instructions"
	"github.com/base/mcm-go/pkg/proposal"
	"github.com/base/mcm-go/pkg/tx"

	"github.com/ethereum/go-ethereum/common"
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
	ProposalWithRoot *proposal.ProposalWithRoot
	TotalSignatures  uint8
}

// InitSignatures initializes the signature storage account
func (s *SignaturesService) InitSignatures(ctx context.Context, params InitSignaturesParams) (solana.Signature, error) {
	proposalWithRoot := params.ProposalWithRoot
	totalSignatures := params.TotalSignatures

	ix, err := instructions.InitSignatures(instructions.InitSignaturesParams{
		MultisigID:      proposalWithRoot.MultisigID,
		Root:            proposalWithRoot.Root,
		ValidUntil:      proposalWithRoot.ValidUntil,
		TotalSignatures: totalSignatures,
		Authority:       s.client.Payer().PublicKey(),
		ProgramID:       s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build init signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS(), s.client.Payer()).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// AppendSignaturesParams contains parameters for appending signatures
type AppendSignaturesParams struct {
	ProposalToSign *proposal.ProposalToSign
	Signatures     [][65]byte // ECDSA signatures in raw format (R + S + V)
}

// AppendSignatures appends a batch of signatures to the storage
func (s *SignaturesService) AppendSignatures(ctx context.Context, params AppendSignaturesParams) (solana.Signature, error) {
	proposalToSign := params.ProposalToSign
	signatures := params.Signatures

	// Sort signatures by recovered EVM address (strictly increasing order)
	sortedSignatures, err := sortSignaturesByAddress(crypto.MessageHash(proposalToSign.HashToSign[:]), signatures)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sort signatures: %w", err)
	}

	// Convert sorted [65]byte signatures to bindings.Signature format
	signaturesBatch := make([]bindings.Signature, len(sortedSignatures))
	for i, sig := range sortedSignatures {
		var r, s [32]byte
		copy(r[:], sig[0:32])
		copy(s[:], sig[32:64])
		signaturesBatch[i] = bindings.Signature{
			V: sig[64],
			R: r,
			S: s,
		}
	}

	ix, err := instructions.AppendSignatures(instructions.AppendSignaturesParams{
		MultisigID:      proposalToSign.MultisigID,
		Root:            proposalToSign.Root,
		ValidUntil:      proposalToSign.ValidUntil,
		SignaturesBatch: signaturesBatch,
		Authority:       s.client.Payer().PublicKey(),
		ProgramID:       s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build append signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS(), s.client.Payer()).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// FinalizeSignaturesParams contains parameters for finalizing signatures
type FinalizeSignaturesParams struct {
	ProposalWithRoot *proposal.ProposalWithRoot
}

// FinalizeSignatures finalizes the signature configuration
func (s *SignaturesService) FinalizeSignatures(ctx context.Context, params FinalizeSignaturesParams) (solana.Signature, error) {
	proposalWithRoot := params.ProposalWithRoot

	ix, err := instructions.FinalizeSignatures(instructions.FinalizeSignaturesParams{
		MultisigID: proposalWithRoot.MultisigID,
		Root:       proposalWithRoot.Root,
		ValidUntil: proposalWithRoot.ValidUntil,
		Authority:  s.client.Payer().PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build finalize signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS(), s.client.Payer()).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// ClearSignaturesParams contains parameters for clearing signatures
type ClearSignaturesParams struct {
	ProposalWithRoot *proposal.ProposalWithRoot
}

// ClearSignatures clears all signatures from the storage
func (s *SignaturesService) ClearSignatures(ctx context.Context, params ClearSignaturesParams) (solana.Signature, error) {
	proposalWithRoot := params.ProposalWithRoot

	ix, err := instructions.ClearSignatures(instructions.ClearSignaturesParams{
		MultisigID: proposalWithRoot.MultisigID,
		Root:       proposalWithRoot.Root,
		ValidUntil: proposalWithRoot.ValidUntil,
		Authority:  s.client.Payer().PublicKey(),
		ProgramID:  s.client.ProgramID,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build clear signatures instruction: %w", err)
	}

	return tx.NewTxBuilder(s.client.RPC, s.client.WS(), s.client.Payer()).
		AddInstruction(ix).
		BuildSignAndSendWithConfirmation(ctx)
}

// signatureWithAddress pairs a signature with its recovered EVM address
type signatureWithAddress struct {
	signature [65]byte
	address   common.Address
}

// sortSignaturesByAddress sorts signatures by their recovered EVM addresses in strictly increasing order
func sortSignaturesByAddress(signedHash [32]byte, signatures [][65]byte) ([][65]byte, error) {
	if len(signatures) == 0 {
		return signatures, nil
	}

	// Create pairs of signature + recovered address
	pairs := make([]signatureWithAddress, len(signatures))

	for i, sig := range signatures {
		// Recover address from signature
		address, err := crypto.RecoverAddressFromSig(signedHash, sig)
		if err != nil {
			return nil, fmt.Errorf("failed to recover address from signature %d: %w", i, err)
		}

		pairs[i] = signatureWithAddress{
			signature: sig,
			address:   address,
		}
	}

	// Sort by address in ascending order
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].address[:], pairs[j].address[:]) < 0
	})

	// Validate strictly increasing order
	for i := 1; i < len(pairs); i++ {
		if bytes.Compare(pairs[i-1].address[:], pairs[i].address[:]) >= 0 {
			return nil, fmt.Errorf("duplicate or non-increasing addresses detected")
		}
	}

	// Extract sorted signatures
	sorted := make([][65]byte, len(signatures))
	for i, pair := range pairs {
		sorted[i] = pair.signature
	}

	return sorted, nil
}
