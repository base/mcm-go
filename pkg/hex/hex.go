// Package hex provides utilities for decoding and parsing hexadecimal strings.
package hex

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Decode decodes a hex string, supporting both with and without "0x" or "0X" prefix.
// This is a general-purpose helper for decoding hex strings in a flexible way.
func Decode(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	return hex.DecodeString(s)
}

// Parse32 parses a hex string into a [32]byte array.
// Supports hex strings with or without "0x"/"0X" prefix.
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
// Supports hex strings with or without "0x"/"0X" prefix.
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
