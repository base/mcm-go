package signers

import (
	"fmt"
	"strconv"
	"strings"
)

// parseUint8Slice parses a comma-separated string into []uint8
func parseUint8Slice(s string) ([]byte, error) {
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

// parseAndPadUint8Array32 parses a comma-separated string and pads it to exactly 32 elements with zeros
func parseAndPadUint8Array32(s string) ([32]uint8, error) {
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
