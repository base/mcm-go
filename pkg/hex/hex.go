// Package hex provides utilities for decoding and parsing hexadecimal strings.
package hex

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Decode decodes a hex string, requiring "0x" prefix.
// This ensures consistent hex string formatting across the CLI.
func Decode(s string) ([]byte, error) {
	if !strings.HasPrefix(s, "0x") {
		return nil, fmt.Errorf("hex string must start with '0x' prefix")
	}
	s = strings.TrimPrefix(s, "0x")
	return hex.DecodeString(s)
}

// Parse32 parses a hex string into a [32]byte array.
// Requires hex strings with "0x" prefix.
func Parse32(s string) ([32]byte, error) {
	var result [32]byte

	data, err := Decode(s)
	if err != nil {
		return result, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(data) != 32 {
		return result, fmt.Errorf("expected 32 bytes, got %d", len(data))
	}

	copy(result[:], data)
	return result, nil
}

// Parse20 parses a hex string into a [20]byte array (EVM address).
// Requires hex strings with "0x" prefix.
func Parse20(s string) ([20]byte, error) {
	var result [20]byte

	data, err := Decode(s)
	if err != nil {
		return result, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(data) != 20 {
		return result, fmt.Errorf("expected 20 bytes (EVM address), got %d", len(data))
	}

	copy(result[:], data)
	return result, nil
}

// ParseSignature parses a single ECDSA signature to [65]byte.
// Format: 0x<130 hex chars> representing r(32) + s(32) + v(1).
func ParseSignature(s string) ([65]byte, error) {
	var result [65]byte

	s = strings.TrimSpace(s)
	data, err := Decode(s)
	if err != nil {
		return result, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(data) != 65 {
		return result, fmt.Errorf("expected 65 bytes, got %d", len(data))
	}

	copy(result[:], data)
	return result, nil
}
