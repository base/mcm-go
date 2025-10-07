// Package client provides a high-level RPC client for interacting with the MCM program.
package client

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Config struct {
	RPCURL    string
	WSURL     string
	ProgramID solana.PublicKey
	Payer     *solana.PrivateKey
}

type Client struct {
	RPC       *rpc.Client
	ws        *ws.Client
	ProgramID solana.PublicKey
	payer     *solana.PrivateKey
}

func New(cfg Config) (*Client, error) {
	rpcClient := rpc.New(cfg.RPCURL)

	var wsClient *ws.Client
	var err error
	if cfg.WSURL != "" {
		wsClient, err = ws.Connect(context.Background(), cfg.WSURL)
		if err != nil {
			return nil, err
		}
	}

	client := &Client{
		RPC:       rpcClient,
		ws:        wsClient,
		ProgramID: cfg.ProgramID,
		payer:     cfg.Payer,
	}

	return client, nil
}

// Payer returns the payer private key, panicking if not set
// This ensures operations that require a wallet fail fast with a clear error
func (c *Client) Payer() solana.PrivateKey {
	if c.payer == nil {
		panic("operation requires a payer (wallet), but none was configured")
	}
	return *c.payer
}

// WS returns the WebSocket client, panicking if not set
// This ensures operations that require transaction confirmations fail fast with a clear error
func (c *Client) WS() *ws.Client {
	if c.ws == nil {
		panic("operation requires a WebSocket connection, but none was configured")
	}
	return c.ws
}

func (c *Client) Close() {
	if c.ws != nil {
		c.ws.Close()
	}
}

// BuildSignAndSend creates a transaction with the given instructions, signs it, and sends it.
// Automatically fetches the latest blockhash and uses the configured payer.
func (c *Client) BuildSignAndSend(ctx context.Context, instructions ...solana.Instruction) (solana.Signature, error) {
	if len(instructions) == 0 {
		return solana.Signature{}, fmt.Errorf("at least one instruction is required")
	}

	// Fetch latest blockhash
	recent, err := c.RPC.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	// Build transaction
	builder := solana.NewTransactionBuilder()
	for _, ix := range instructions {
		builder = builder.AddInstruction(ix)
	}
	tx, err := builder.
		SetRecentBlockHash(recent.Value.Blockhash).
		SetFeePayer(c.Payer().PublicKey()).
		Build()
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(c.Payer().PublicKey()) {
			payer := c.Payer()
			return &payer
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	sig, err := c.RPC.SendTransaction(ctx, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

// BuildSignAndSendWithConfirmation creates, signs, sends, and confirms a transaction.
// Automatically fetches the latest blockhash and uses the configured payer.
// Requires WebSocket client to be configured.
func (c *Client) BuildSignAndSendWithConfirmation(ctx context.Context, instructions ...solana.Instruction) (solana.Signature, error) {
	if len(instructions) == 0 {
		return solana.Signature{}, fmt.Errorf("at least one instruction is required")
	}

	// Fetch latest blockhash
	recent, err := c.RPC.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	// Build transaction
	builder := solana.NewTransactionBuilder()
	for _, ix := range instructions {
		builder = builder.AddInstruction(ix)
	}
	tx, err := builder.
		SetRecentBlockHash(recent.Value.Blockhash).
		SetFeePayer(c.Payer().PublicKey()).
		Build()
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(c.Payer().PublicKey()) {
			payer := c.Payer()
			return &payer
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send and confirm transaction
	sig, err := confirm.SendAndConfirmTransaction(ctx, c.RPC, c.WS(), tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send and confirm transaction: %w", err)
	}

	return sig, nil
}
