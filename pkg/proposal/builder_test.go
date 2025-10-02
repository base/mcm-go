package proposal

import (
	"encoding/hex"
	"testing"

	"mcm-go/pkg/crypto"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
)

// Test values from Rust test suite
// Last 8 bytes of keccak256("solana:localnet") as big-endian
// This is 0x4808e31713a26612 --> in little-endian, it is "1266a21317e30848"
const testChainID uint64 = 5190648258797659666

func decode32(s string) [32]byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	var result [32]byte
	copy(result[:], data)
	return result
}

// TestEncodeMetadataLeaf verifies that Go implementation matches Rust test
// From eth_utils.rs test_hash_leaf::valid (lines 417-439)
func TestEncodeMetadataLeaf(t *testing.T) {
	multisigBytes := decode32("b870e12dd379891561d2e9fa8f26431834eb736f2f24fc2a2a4dff1fd5dca4df")
	metadata := RootMetadata{
		ChainID:              testChainID,
		Multisig:             solana.PublicKeyFromBytes(multisigBytes[:]),
		PreOpCount:           0,
		PostOpCount:          1,
		OverridePreviousRoot: false,
	}

	result, err := encodeMetadataLeaf(metadata)
	assert.NoError(t, err)

	expected := decode32("c31496e313ba769f8c9f061dd35c6aa06c1c51ec9111f54be7be307a0de6b556")
	assert.Equal(t, expected, result, "Metadata leaf hash must match Rust implementation")

	// Verify the encoding format matches Rust comment (lines 433-438)
	// 47fded70901d27038394db905a72563cad6f07581dbcdd1472ccd2f742af6360 <-- METADATA_DOMAIN_SEPARATOR
	// 0000000000000000000000000000000000000000000000001266a21317e30848 <-- chain_id (LE left-padded)
	// b870e12dd379891561d2e9fa8f26431834eb736f2f24fc2a2a4dff1fd5dca4df <-- multisig
	// 0000000000000000000000000000000000000000000000000000000000000000 <-- pre_op_count
	// 0000000000000000000000000000000000000000000000000100000000000000 <-- post_op_count (1 in LE left-padded)
	// 0000000000000000000000000000000000000000000000000000000000000000 <-- override_previous_root
}

// TestEncodeOperationLeaf_Empty verifies empty op matches Rust test
// From execute.rs test_op_hash_leaf::empty_op (lines 224-250)
func TestEncodeOperationLeaf_Empty(t *testing.T) {
	multisigBytes := decode32("eceeab9f961bbf0050babcbf22663ee37905749830f41cd5096fbc9b158f8b13")
	toBytes := decode32("30d721519466e7a7c60691f75f49734954bee02fdb6700dd7f02a19de972ccdb")
	multisig := solana.PublicKeyFromBytes(multisigBytes[:])
	to := solana.PublicKeyFromBytes(toBytes[:])
	data, _ := hex.DecodeString("d62c04f70c29d96e")

	result, err := encodeOperationLeaf(
		testChainID,
		multisig,
		0, // nonce
		to,
		data,
		[]*solana.AccountMeta{}, // empty remaining accounts
	)
	assert.NoError(t, err)

	expected := decode32("0145195c134f5cec64fba146648763c2a7ac9bdb2c2efbc31a72bf9fb5f4246c")
	assert.Equal(t, expected, result, "Empty operation leaf hash must match Rust implementation")

	// Verify the encoding format matches Rust comment (lines 238-246)
	// fb98816ff3c5138a68abfd40b8d8fbc22972fea1dd89757331327e6e0a9440b7  separator
	// 0000000000000000000000000000000000000000000000001266a21317e30848  chain_id
	// eceeab9f961bbf0050babcbf22663ee37905749830f41cd5096fbc9b158f8b13  multisig
	// 0000000000000000000000000000000000000000000000000000000000000000  nonce
	// 30d721519466e7a7c60691f75f49734954bee02fdb6700dd7f02a19de972ccdb  to
	// 0000000000000000000000000000000000000000000000000800000000000000  data_len (8 in LE left-padded)
	// d62c04f70c29d96e                                                  data
	// 0000000000000000000000000000000000000000000000000000000000000000  remaining_accounts_len
}

// TestEncodeOperationLeaf_WithData verifies op with data matches Rust test
// From execute.rs test_op_hash_leaf::u8_instruction_data (lines 254-282)
func TestEncodeOperationLeaf_WithData(t *testing.T) {
	multisigBytes := decode32("eceeab9f961bbf0050babcbf22663ee37905749830f41cd5096fbc9b158f8b13")
	toBytes := decode32("30d721519466e7a7c60691f75f49734954bee02fdb6700dd7f02a19de972ccdb")
	multisig := solana.PublicKeyFromBytes(multisigBytes[:])
	to := solana.PublicKeyFromBytes(toBytes[:])
	data, _ := hex.DecodeString("11af9cfd5bad1ae47b")

	result, err := encodeOperationLeaf(
		testChainID,
		multisig,
		1, // nonce
		to,
		data,
		[]*solana.AccountMeta{}, // empty remaining accounts
	)
	assert.NoError(t, err)

	expected := decode32("d0dade6731524fbf07bf67fb9d250bc10b6e22adcd5e04ad9d6f515c75ac4951")
	assert.Equal(t, expected, result, "Operation leaf with data hash must match Rust implementation")

	// Verify the encoding format matches Rust comment (lines 268-276)
	// Buffer[0]: fb98816ff3c5138a68abfd40b8d8fbc22972fea1dd89757331327e6e0a9440b7
	// Buffer[1]: 0000000000000000000000000000000000000000000000001266a21317e30848
	// Buffer[2]: eceeab9f961bbf0050babcbf22663ee37905749830f41cd5096fbc9b158f8b13
	// Buffer[3]: 0000000000000000000000000000000000000000000000000100000000000000  <-- nonce=1 LE left-padded
	// Buffer[4]: 30d721519466e7a7c60691f75f49734954bee02fdb6700dd7f02a19de972ccdb
	// Buffer[5]: 0000000000000000000000000000000000000000000000000900000000000000  <-- data_len=9 LE left-padded
	// Buffer[6]: 11af9cfd5bad1ae47b
	// Buffer[7]: 0000000000000000000000000000000000000000000000000000000000000000
}

// TestDomainSeparators verifies domain separators match Rust constants
func TestDomainSeparators(t *testing.T) {
	// From eth_utils.rs lines 34-36
	expectedMetadata := decode32("47fded70901d27038394db905a72563cad6f07581dbcdd1472ccd2f742af6360")
	assert.Equal(t, expectedMetadata, domainSeparatorMetadata, "Metadata domain separator must match Rust")

	// Verify it's keccak256("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_METADATA_SOLANA")
	computed := crypto.Keccak256Hash([]byte("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_METADATA_SOLANA"))
	assert.Equal(t, expectedMetadata, computed, "Metadata domain separator should be keccak256 of the constant string")

	// From eth_utils.rs lines 43-45
	expectedOp := decode32("fb98816ff3c5138a68abfd40b8d8fbc22972fea1dd89757331327e6e0a9440b7")
	assert.Equal(t, expectedOp, domainSeparatorOp, "Op domain separator must match Rust")

	// Verify it's keccak256("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_OP_SOLANA")
	computed = crypto.Keccak256Hash([]byte("MANY_CHAIN_MULTI_SIG_DOMAIN_SEPARATOR_OP_SOLANA"))
	assert.Equal(t, expectedOp, computed, "Op domain separator should be keccak256 of the constant string")
}

// TestComputeHashToSign verifies the hash computation for signer validation
// This hash is: keccak256(root || validUntil_as_4_bytes_LE)
func TestComputeHashToSign(t *testing.T) {
	root := decode32("d5ef592d1ad183db43b4980d7ab7ee43a6f6a284988c3e3a23d38c07beb520c7")
	validUntil := uint32(1748317727)

	result := computeHashToSign(root, validUntil)

	// Verify deterministic behavior
	result2 := computeHashToSign(root, validUntil)
	assert.Equal(t, result, result2, "Hash must be deterministic")

	// Verify different inputs produce different hashes
	differentRoot := decode32("0000000000000000000000000000000000000000000000000000000000000000")
	differentResult := computeHashToSign(differentRoot, validUntil)
	assert.NotEqual(t, result, differentResult, "Different roots must produce different hashes")

	differentValidUntil := uint32(1748317728)
	differentResult2 := computeHashToSign(root, differentValidUntil)
	assert.NotEqual(t, result, differentResult2, "Different validUntil must produce different hashes")
}
