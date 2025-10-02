// Package client provides a high-level RPC client for interacting with the MCM program.
package client

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Client wraps the Solana RPC client with MCM-specific functionality
type Client struct {
	RPC          *rpc.Client
	WS           *ws.Client
	ProgramID    solana.PublicKey
	DefaultPayer *solana.PrivateKey
}

// Config contains configuration for creating a new MCM client
type Config struct {
	RPCURL       string
	WSURL        string
	ProgramID    solana.PublicKey
	DefaultPayer *solana.PrivateKey
}

// New creates a new MCM client with the given configuration
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
		RPC:          rpcClient,
		WS:           wsClient,
		ProgramID:    cfg.ProgramID,
		DefaultPayer: cfg.DefaultPayer,
	}

	return client, nil
}

// Close closes the WebSocket connection if it exists
func (c *Client) Close() {
	if c.WS != nil {
		c.WS.Close()
	}
}

// GetPayer returns the default payer or panics if not set
func (c *Client) GetPayer() solana.PublicKey {
	if c.DefaultPayer == nil {
		panic("default payer not set")
	}
	return c.DefaultPayer.PublicKey()
}
