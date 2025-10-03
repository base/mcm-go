package cli

import "testing"

func TestResolveNetworkAlias(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		isWebSocket bool
		expected    string
	}{
		// Mainnet
		{"mainnet RPC", "mainnet", false, "https://api.mainnet-beta.solana.com"},
		{"mainnet WS", "mainnet", true, "wss://api.mainnet-beta.solana.com"},
		{"mainnet-beta RPC", "mainnet-beta", false, "https://api.mainnet-beta.solana.com"},
		{"mainnet-beta WS", "mainnet-beta", true, "wss://api.mainnet-beta.solana.com"},

		// Devnet
		{"devnet RPC", "devnet", false, "https://api.devnet.solana.com"},
		{"devnet WS", "devnet", true, "wss://api.devnet.solana.com"},

		// Testnet
		{"testnet RPC", "testnet", false, "https://api.testnet.solana.com"},
		{"testnet WS", "testnet", true, "wss://api.testnet.solana.com"},

		// Localhost
		{"localhost RPC", "localhost", false, "http://localhost:8899"},
		{"localhost WS", "localhost", true, "ws://localhost:8900"},
		{"local RPC", "local", false, "http://localhost:8899"},
		{"local WS", "local", true, "ws://localhost:8900"},

		// Custom URLs (passthrough)
		{"custom https", "https://custom.rpc.com", false, "https://custom.rpc.com"},
		{"custom wss", "wss://custom.ws.com", true, "wss://custom.ws.com"},
		{"custom http", "http://192.168.1.100:8899", false, "http://192.168.1.100:8899"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveNetworkAlias(tt.url, tt.isWebSocket)
			if result != tt.expected {
				t.Errorf("resolveNetworkAlias(%q, %v) = %q, want %q", tt.url, tt.isWebSocket, result, tt.expected)
			}
		})
	}
}
