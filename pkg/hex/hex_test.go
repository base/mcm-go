package hex

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "without prefix",
			input:   "deadbeef",
			want:    "deadbeef",
			wantErr: false,
		},
		{
			name:    "with lowercase 0x",
			input:   "0xdeadbeef",
			want:    "deadbeef",
			wantErr: false,
		},
		{
			name:    "with uppercase 0X",
			input:   "0XDEADBEEF",
			want:    "deadbeef",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid hex",
			input:   "0xzzzz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, hex.EncodeToString(got))
			}
		})
	}
}

func TestParse32(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid hex without prefix",
			input:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr: false,
		},
		{
			name:    "valid hex with lowercase 0x prefix",
			input:   "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr: false,
		},
		{
			name:    "valid hex with uppercase 0X prefix",
			input:   "0X0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
			wantErr: false,
		},
		{
			name:    "too short",
			input:   "0123456789abcdef",
			wantErr: true,
		},
		{
			name:    "invalid hex",
			input:   "xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 32, len(result))
			}
		})
	}
}

func TestParse20(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid EVM address without prefix",
			input:   "1234567890abcdef1234567890abcdef12345678",
			wantErr: false,
		},
		{
			name:    "valid EVM address with lowercase 0x prefix",
			input:   "0x1234567890abcdef1234567890abcdef12345678",
			wantErr: false,
		},
		{
			name:    "valid EVM address with uppercase 0X prefix",
			input:   "0X1234567890ABCDEF1234567890ABCDEF12345678",
			wantErr: false,
		},
		{
			name:    "too short",
			input:   "1234567890",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "1234567890abcdef1234567890abcdef1234567890",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse20(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 20, len(result))
			}
		})
	}
}
