// Package tx provides utilities for building and submitting Solana transactions.
package tx

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Builder helps construct Solana transactions with proper recent blockhash and signing
type Builder struct {
	rpc          *rpc.Client
	ws           *ws.Client
	instructions []solana.Instruction
	signers      []solana.PrivateKey
	payer        solana.PrivateKey
}

// NewTxBuilder creates a new transaction builder
func NewTxBuilder(rpcClient *rpc.Client, wsClient *ws.Client, payer solana.PrivateKey) *Builder {
	return &Builder{
		rpc:          rpcClient,
		ws:           wsClient,
		instructions: make([]solana.Instruction, 0),
		signers:      []solana.PrivateKey{payer},
		payer:        payer,
	}
}

// AddInstruction adds an instruction to the transaction
func (b *Builder) AddInstruction(ix solana.Instruction) *Builder {
	b.instructions = append(b.instructions, ix)
	return b
}

// AddSigner adds an additional signer to the transaction (payer is already included)
func (b *Builder) AddSigner(signer solana.PrivateKey) *Builder {
	b.signers = append(b.signers, signer)
	return b
}

// Build creates a transaction with a recent blockhash
func (b *Builder) Build(ctx context.Context) (*solana.Transaction, error) {
	if len(b.instructions) == 0 {
		return nil, fmt.Errorf("no instructions added to transaction")
	}

	// Get latest blockhash (replaces deprecated GetRecentBlockhash)
	recent, err := b.rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	// Build transaction
	tx, err := solana.NewTransaction(
		b.instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(b.payer.PublicKey()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

// BuildAndSign creates and signs a transaction
func (b *Builder) BuildAndSign(ctx context.Context) (*solana.Transaction, error) {
	tx, err := b.Build(ctx)
	if err != nil {
		return nil, err
	}

	// Sign transaction
	if len(b.signers) == 0 {
		return nil, fmt.Errorf("no signers provided")
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, signer := range b.signers {
			if signer.PublicKey().Equals(key) {
				return &signer
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// BuildSignAndSend creates, signs, and sends a transaction
func (b *Builder) BuildSignAndSend(ctx context.Context) (solana.Signature, error) {
	tx, err := b.BuildAndSign(ctx)
	if err != nil {
		return solana.Signature{}, err
	}

	sig, err := b.rpc.SendTransaction(ctx, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

// BuildSignAndSendWithConfirmation creates, signs, sends, and confirms a transaction
func (b *Builder) BuildSignAndSendWithConfirmation(ctx context.Context) (solana.Signature, error) {
	tx, err := b.BuildAndSign(ctx)
	if err != nil {
		return solana.Signature{}, err
	}

	// Use solana-go's built-in confirmation mechanism
	if b.ws == nil {
		return solana.Signature{}, fmt.Errorf("websocket client required for confirmation")
	}

	sig, err := confirm.SendAndConfirmTransaction(ctx, b.rpc, b.ws, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send and confirm transaction: %w", err)
	}

	return sig, nil
}

// Simulate simulates the transaction without sending it
func (b *Builder) Simulate(ctx context.Context) (*rpc.SimulateTransactionResponse, error) {
	tx, err := b.BuildAndSign(ctx)
	if err != nil {
		return nil, err
	}

	// Simulate with options for better accuracy
	result, err := b.rpc.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		SigVerify:  true,
		Commitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	return result, nil
}

// Reset clears all instructions
func (b *Builder) Reset() *Builder {
	b.instructions = make([]solana.Instruction, 0)
	b.signers = []solana.PrivateKey{b.payer}
	return b
}
