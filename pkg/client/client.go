// Package client provides a high-level RPC client for interacting with the MCM program.
package client

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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
