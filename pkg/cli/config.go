package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/base/mcm-go/pkg/client"

	"github.com/gagliardetto/solana-go"
)

// ConfigParams holds parameters for loading config
type ConfigParams struct {
	RPCUrl      string
	WSUrl       string
	ProgramID   string
	KeypairPath string
}

// LoadConfig loads configuration from flags/env and returns a client.Config
func LoadConfig(params ConfigParams) (*client.Config, error) {
	// Validate RPC URL
	if params.RPCUrl == "" {
		return nil, fmt.Errorf("RPC URL is required (--rpc-url or MCM_RPC_URL)")
	}

	// Resolve network aliases to actual endpoints
	rpcURL := resolveNetworkAlias(params.RPCUrl, false)

	// WS URL is optional (only needed for transactions)
	wsURL := ""
	if params.WSUrl != "" {
		wsURL = resolveNetworkAlias(params.WSUrl, true)
	}

	// Parse program ID
	programID, err := solana.PublicKeyFromBase58(params.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	// Keypair is optional (only needed for transactions)
	var payer *solana.PrivateKey
	if params.KeypairPath != "" {
		key, err := loadKeypair(params.KeypairPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keypair: %w", err)
		}
		payer = &key
	}

	return &client.Config{
		RPCURL:    rpcURL,
		WSURL:     wsURL,
		ProgramID: programID,
		Payer:     payer,
	}, nil
}

// DefaultKeypairPath returns the default Solana keypair path
func DefaultKeypairPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "solana", "id.json")
}

// resolveNetworkAlias converts network aliases to actual endpoints
func resolveNetworkAlias(url string, isWebSocket bool) string {
	switch url {
	case "mainnet", "mainnet-beta":
		if isWebSocket {
			return "wss://api.mainnet-beta.solana.com"
		}
		return "https://api.mainnet-beta.solana.com"
	case "devnet":
		if isWebSocket {
			return "wss://api.devnet.solana.com"
		}
		return "https://api.devnet.solana.com"
	case "testnet":
		if isWebSocket {
			return "wss://api.testnet.solana.com"
		}
		return "https://api.testnet.solana.com"
	case "localhost", "local":
		if isWebSocket {
			return "ws://localhost:8900"
		}
		return "http://localhost:8899"
	default:
		return url
	}
}

// loadKeypair loads a keypair from file (JSON array or base58 string)
func loadKeypair(path string) (solana.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return solana.PrivateKey{}, fmt.Errorf("failed to read keypair file: %w", err)
	}

	// Try JSON array format first (solana-keygen format)
	var jsonKey []byte
	if err := json.Unmarshal(data, &jsonKey); err == nil {
		if len(jsonKey) != 64 {
			return solana.PrivateKey{}, fmt.Errorf("invalid keypair length: expected 64 bytes, got %d", len(jsonKey))
		}
		return solana.PrivateKey(jsonKey), nil
	}

	// Try base58 format
	key, err := solana.PrivateKeyFromBase58(string(data))
	if err != nil {
		return solana.PrivateKey{}, fmt.Errorf("failed to parse keypair (tried JSON and base58): %w", err)
	}

	return key, nil
}
