package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"mcm-go/pkg/hex"
)

// ParseHex32 parses a hex string into a [32]byte array.
// Deprecated: Use hex.Parse32 directly instead.
func ParseHex32(s string) ([32]byte, error) {
	return hex.Parse32(s)
}

// ParseHex20 parses a hex string into a [20]byte array (EVM address).
// Deprecated: Use hex.Parse20 directly instead.
func ParseHex20(s string) ([20]byte, error) {
	return hex.Parse20(s)
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
