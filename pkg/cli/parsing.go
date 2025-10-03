package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// ParseHex32 parses a hex string into a [32]byte array
func ParseHex32(s string) ([32]byte, error) {
	var result [32]byte

	// Remove 0x prefix if present
	s = strings.TrimPrefix(s, "0x")

	// Decode hex
	data, err := hex.DecodeString(s)
	if err != nil {
		return result, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(data) != 32 {
		return result, fmt.Errorf("expected 32 bytes, got %d", len(data))
	}

	copy(result[:], data)
	return result, nil
}

// ParseHex20 parses a hex string into a [20]byte array (EVM address)
func ParseHex20(s string) ([20]byte, error) {
	var result [20]byte

	// Remove 0x prefix if present
	s = strings.TrimPrefix(s, "0x")

	// Decode hex
	data, err := hex.DecodeString(s)
	if err != nil {
		return result, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(data) != 20 {
		return result, fmt.Errorf("expected 20 bytes (EVM address), got %d", len(data))
	}

	copy(result[:], data)
	return result, nil
}

// ParseAndSortSigners parses a comma-separated list of hex signers,
// validates them, sorts them in strictly increasing order, and checks for duplicates.
// Returns the sorted signers and a boolean indicating if sorting occurred.
func ParseAndSortSigners(signersStr string) ([][20]uint8, bool, error) {
	// Split by comma
	parts := strings.Split(signersStr, ",")
	if len(parts) == 0 {
		return nil, false, fmt.Errorf("empty signers list")
	}

	// Parse each signer
	signers := make([][20]uint8, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		signer, err := ParseHex20(part)
		if err != nil {
			return nil, false, fmt.Errorf("invalid signer at index %d: %w", i, err)
		}

		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		return nil, false, fmt.Errorf("no valid signers provided")
	}

	// Check for duplicates before sorting
	seen := make(map[[20]byte]bool)
	for _, s := range signers {
		if seen[s] {
			return nil, false, fmt.Errorf("duplicate signer address: 0x%x", s)
		}
		seen[s] = true
	}

	// Keep original for comparison
	original := make([][20]uint8, len(signers))
	copy(original, signers)

	// Sort in lexicographically increasing order
	sort.Slice(signers, func(i, j int) bool {
		return bytes.Compare(signers[i][:], signers[j][:]) < 0
	})

	// Check if sorting changed the order
	sorted := false
	for i := range signers {
		if signers[i] != original[i] {
			sorted = true
			break
		}
	}

	return signers, sorted, nil
}
