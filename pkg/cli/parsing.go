package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/base/mcm-go/pkg/hex"
	"github.com/gagliardetto/solana-go"
)

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

		signer, err := hex.Parse20(part)
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

// ParseUint8Slice parses a comma-separated string into []uint8
func ParseUint8Slice(s string) ([]byte, error) {
	if s == "" {
		return nil, fmt.Errorf("empty input")
	}

	parts := strings.Split(s, ",")
	result := make([]byte, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid value at index %d: %s", i, part)
		}
		result[i] = uint8(val)
	}

	return result, nil
}

// ParseAndPadUint8Array32 parses a comma-separated string and pads it to exactly 32 elements with zeros
func ParseAndPadUint8Array32(s string) ([32]uint8, error) {
	var result [32]uint8

	if s == "" {
		return result, fmt.Errorf("empty input")
	}

	parts := strings.Split(s, ",")
	if len(parts) > 32 {
		return result, fmt.Errorf("too many values: got %d, max 32", len(parts))
	}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return result, fmt.Errorf("invalid value at index %d: %s", i, part)
		}
		result[i] = uint8(val)
	}

	// Remaining elements are already 0 due to zero-initialization
	return result, nil
}

// ParseProgramID parses a program ID from a base58 string
func ParseProgramID(programIDStr string) (solana.PublicKey, error) {
	programID, err := solana.PublicKeyFromBase58(programIDStr)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("invalid program ID: %w", err)
	}
	return programID, nil
}
