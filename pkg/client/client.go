// Package client provides a high-level RPC client for interacting with the MCM program.
package client

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Client struct {
	RPC       *rpc.Client
	WS        *ws.Client
	ProgramID solana.PublicKey
	Payer     solana.PrivateKey
}

type Config struct {
	RPCURL    string
	WSURL     string
	ProgramID solana.PublicKey
	Payer     solana.PrivateKey
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
		WS:        wsClient,
		ProgramID: cfg.ProgramID,
		Payer:     cfg.Payer,
	}

	return client, nil
}

func (c *Client) Close() {
	if c.WS != nil {
		c.WS.Close()
	}
}
