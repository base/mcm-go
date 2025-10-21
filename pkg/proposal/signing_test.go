package proposal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEIP712DomainStructure verifies the EIP-712 domain separator structure
func TestEIP712DomainStructure(t *testing.T) {
	root := [32]byte{1, 2, 3}
	validUntil := uint32(123456789)
	chainID := uint64(5190648258797659666)
	programID := [32]byte{4, 5, 6}

	hash, domainSep, structHash, err := computeEIP712HashToSign(root, validUntil, chainID, programID)
	require.NoError(t, err, "Failed to compute hash")
	assert.NotEqual(t, [32]byte{}, domainSep, "DomainSeparator should not be zero")
	assert.NotEqual(t, [32]byte{}, structHash, "StructHash should not be zero")

	// Verify that different inputs produce different hashes
	hash2, _, _, err := computeEIP712HashToSign(root, validUntil+1, chainID, programID)
	require.NoError(t, err, "Failed to compute hash2")

	assert.NotEqual(t, hash, hash2, "Different validUntil should produce different hashes")

	hash3, _, _, err := computeEIP712HashToSign([32]byte{7, 8, 9}, validUntil, chainID, programID)
	require.NoError(t, err, "Failed to compute hash3")

	assert.NotEqual(t, hash, hash3, "Different root should produce different hashes")
}

// TestWithHashToSignIntegration tests the full integration with ProposalWithRoot
func TestWithHashToSignIntegration(t *testing.T) {
	// Create a minimal proposal with root
	multisigID := [32]byte{1}
	validUntil := uint32(123456789)
	chainID := uint64(5190648258797659666)

	pwr := &ProposalWithRoot{
		Proposal: &Proposal{
			MultisigID: multisigID,
			ValidUntil: validUntil,
			RootMetadata: RootMetadata{
				ChainID:              chainID,
				PreOpCount:           0,
				PostOpCount:          1,
				OverridePreviousRoot: false,
			},
		},
		ProposalRoot: ProposalRoot{
			Root: [32]byte{2, 3, 4},
		},
	}

	programID := [32]byte{5, 6, 7}
	var programIDPubkey [32]byte
	copy(programIDPubkey[:], programID[:])

	// Compute hash to sign
	pts, err := pwr.WithHashToSign(programIDPubkey)
	require.NoError(t, err, "Failed to compute hash to sign")

	// Verify the ProposalToSign structure
	assert.NotNil(t, pts, "ProposalToSign should not be nil")
	assert.Equal(t, pwr, pts.ProposalWithRoot, "ProposalWithRoot should be preserved")
	assert.NotEqual(t, [32]byte{}, pts.MessageHash, "MessageHash should not be zero")
	assert.NotEqual(t, [32]byte{}, pts.DomainSeparator, "DomainSeparator should not be zero")
	assert.NotEqual(t, [32]byte{}, pts.StructHash, "StructHash should not be zero")

	t.Logf("Message hash: 0x%x", pts.MessageHash)
	t.Logf("Domain separator: 0x%x", pts.DomainSeparator)
	t.Logf("Struct hash: 0x%x", pts.StructHash)
	t.Logf("Combined (for signing tool): 0x%x%x", pts.DomainSeparator, pts.StructHash)
}
