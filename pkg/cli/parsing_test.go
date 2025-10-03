package cli

import (
	"encoding/hex"
	"testing"

	hexPkg "mcm-go/pkg/hex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHex32(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid hex without prefix should fail",
			input:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr: true,
		},
		{
			name:    "valid hex with lowercase 0x prefix",
			input:   "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr: false,
		},
		{
			name:    "valid hex with uppercase 0X prefix should fail",
			input:   "0X0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "0x0123456789abcdef",
			wantErr: true,
		},
		{
			name:    "invalid hex",
			input:   "0xxyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hexPkg.Parse32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 32, len(result))
			}
		})
	}
}

func TestParseHex20(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid EVM address without prefix should fail",
			input:   "1234567890abcdef1234567890abcdef12345678",
			wantErr: true,
		},
		{
			name:    "valid EVM address with lowercase 0x prefix",
			input:   "0x1234567890abcdef1234567890abcdef12345678",
			wantErr: false,
		},
		{
			name:    "valid EVM address with uppercase 0X prefix should fail",
			input:   "0X1234567890ABCDEF1234567890ABCDEF12345678",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "0x1234567890",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "0x1234567890abcdef1234567890abcdef1234567890",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hexPkg.Parse20(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 20, len(result))
			}
		})
	}
}

func TestParseAndSortSigners(t *testing.T) {
	t.Run("sorts signers in increasing order", func(t *testing.T) {
		input := "0x3333333333333333333333333333333333333333,0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222"
		signers, sorted, err := ParseAndSortSigners(input)
		require.NoError(t, err)
		assert.True(t, sorted, "signers should be sorted")
		assert.Equal(t, 3, len(signers))

		// Verify order
		assert.Equal(t, "0x1111111111111111111111111111111111111111", "0x"+hex.EncodeToString(signers[0][:]))
		assert.Equal(t, "0x2222222222222222222222222222222222222222", "0x"+hex.EncodeToString(signers[1][:]))
		assert.Equal(t, "0x3333333333333333333333333333333333333333", "0x"+hex.EncodeToString(signers[2][:]))
	})

	t.Run("already sorted", func(t *testing.T) {
		input := "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222"
		signers, sorted, err := ParseAndSortSigners(input)
		require.NoError(t, err)
		assert.False(t, sorted, "signers were already sorted")
		assert.Equal(t, 2, len(signers))
	})

	t.Run("detects duplicates", func(t *testing.T) {
		input := "0x1111111111111111111111111111111111111111,0x1111111111111111111111111111111111111111"
		_, _, err := ParseAndSortSigners(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	t.Run("empty list", func(t *testing.T) {
		_, _, err := ParseAndSortSigners("")
		assert.Error(t, err)
	})

	t.Run("invalid hex without prefix", func(t *testing.T) {
		input := "1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222"
		_, _, err := ParseAndSortSigners(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "0x")
	})

	t.Run("handles whitespace", func(t *testing.T) {
		input := " 0x1111111111111111111111111111111111111111 , 0x2222222222222222222222222222222222222222 "
		signers, _, err := ParseAndSortSigners(input)
		require.NoError(t, err)
		assert.Equal(t, 2, len(signers))
	})
}
