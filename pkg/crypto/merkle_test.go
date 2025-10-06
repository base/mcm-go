package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func decode32(s string) [32]byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	var result [32]byte
	copy(result[:], data)
	return result
}

// TestHashPair_BasicKeccak verifies basic hash_pair behavior
// From eth_utils.rs test_hash_pair::basic_keccak (lines 288-297)
func TestHashPair_BasicKeccak(t *testing.T) {
	a := [32]byte{}
	b := [32]byte{}
	result := hashPair(a, b)

	expected := decode32("ad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5")
	assert.Equal(t, expected, result, "hash_pair([0;32], [0;32]) must match Rust implementation")
}

// TestHashPair_Ordering verifies hash_pair is commutative (ordering-independent)
// From eth_utils.rs test_hash_pair::ordering (lines 299-309)
func TestHashPair_Ordering(t *testing.T) {
	var a [32]byte // all zeros
	var b [32]byte
	for i := range b {
		b[i] = 1 // all ones
	}

	expected := decode32("d5f4f7e1d989848480236fb0a5f808d5877abf778364ae50845234dd6c1e80fc")

	// Hash should be the same regardless of order
	assert.Equal(t, expected, hashPair(a, b), "hash_pair(a, b) must match Rust implementation")
	assert.Equal(t, expected, hashPair(b, a), "hash_pair(b, a) must equal hash_pair(a, b)")
}

// TestVerifyMerkleProof_Identity verifies single leaf with empty proof
// From eth_utils.rs test_compute_merkle_root::identity (lines 353-361)
func TestVerifyMerkleProof_Identity(t *testing.T) {
	leaf := [32]byte{} // all zeros
	proof := Proof{}   // empty proof
	root := [32]byte{} // all zeros

	// A single leaf with no proof should verify against itself as root
	result := VerifyMerkleProof(leaf, proof, root)
	assert.True(t, result, "Single leaf with empty proof should verify against itself")
}

// TestVerifyMerkleProof_SingleStep verifies proof with one sibling
// From eth_utils.rs test_compute_merkle_root::single_step (lines 363-376)
//
// Tree structure:
//
//	root: 8e4b8e18156a1c7271055ce5b7ef53bb370294ebd631a3b95418a92da46e681f
//	       /                    \
//	leaf: 0000...              sibling: 1111...
func TestVerifyMerkleProof_SingleStep(t *testing.T) {
	leaf := decode32("0000000000000000000000000000000000000000000000000000000000000000")
	sibling := decode32("1111111111111111111111111111111111111111111111111111111111111111")
	root := decode32("8e4b8e18156a1c7271055ce5b7ef53bb370294ebd631a3b95418a92da46e681f")

	proof := Proof{sibling}

	result := VerifyMerkleProof(leaf, proof, root)
	assert.True(t, result, "Proof verification must match Rust calculate_merkle_root")
}

// TestVerifyMerkleProof_MultiStep verifies proof with multiple siblings
// From eth_utils.rs test_compute_merkle_root::multi_step (lines 378-392)
//
// Tree structure (3 leaves):
//
//	                 root: 888aba2887457beba19643fd1c5e5be943d3f0b910d418c1ab49c057c06f6738
//	                         /                                                    \
//	intermediate: 8e4b8e18156a1c7271055ce5b7ef53bb370294ebd631a3b95418a92da46e681f    leaf3: 2222...
//	               /                    \
//	         leaf1: 0000...         leaf2: 1111... (our leaf)
func TestVerifyMerkleProof_MultiStep(t *testing.T) {
	leaf := decode32("1111111111111111111111111111111111111111111111111111111111111111")

	// Proof consists of:
	// 1. Sibling at same level (0000...)
	// 2. Sibling at parent level (2222...)
	proof := Proof{
		decode32("0000000000000000000000000000000000000000000000000000000000000000"),
		decode32("2222222222222222222222222222222222222222222222222222222222222222"),
	}

	root := decode32("888aba2887457beba19643fd1c5e5be943d3f0b910d418c1ab49c057c06f6738")

	result := VerifyMerkleProof(leaf, proof, root)
	assert.True(t, result, "Multi-step proof verification must match Rust calculate_merkle_root")
}

// TestBuildMerkleTreeFromLeaves_ThreeLeaves verifies tree construction with the same 3-leaf example
func TestBuildMerkleTreeFromLeaves_ThreeLeaves(t *testing.T) {
	// Same 3 leaves as in Rust test
	leaves := [][32]byte{
		decode32("0000000000000000000000000000000000000000000000000000000000000000"),
		decode32("1111111111111111111111111111111111111111111111111111111111111111"),
		decode32("2222222222222222222222222222222222222222222222222222222222222222"),
	}

	expectedRoot := decode32("888aba2887457beba19643fd1c5e5be943d3f0b910d418c1ab49c057c06f6738")

	tree, err := BuildMerkleTreeFromLeaves(leaves)
	assert.NoError(t, err)
	assert.Equal(t, expectedRoot, tree.Root, "Tree root must match Rust implementation")

	// Verify all proofs work
	for i, leaf := range leaves {
		valid := VerifyMerkleProof(leaf, tree.Proofs[i], tree.Root)
		assert.True(t, valid, "Proof for leaf %d must verify", i)
	}
}

// TestBuildMerkleTreeFromLeaves_SingleLeaf verifies edge case with one leaf
func TestBuildMerkleTreeFromLeaves_SingleLeaf(t *testing.T) {
	leaf := decode32("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	leaves := [][32]byte{leaf}

	tree, err := BuildMerkleTreeFromLeaves(leaves)
	assert.NoError(t, err)

	// Root should be the leaf itself
	assert.Equal(t, leaf, tree.Root, "Single leaf tree root should equal the leaf")

	// Proof should be empty
	assert.Len(t, tree.Proofs, 1)
	assert.Len(t, tree.Proofs[0], 0, "Single leaf should have empty proof")

	// Should verify
	valid := VerifyMerkleProof(leaf, tree.Proofs[0], tree.Root)
	assert.True(t, valid, "Single leaf with empty proof should verify")
}

// TestBuildMerkleTreeFromLeaves_TwoLeaves verifies tree with two leaves
func TestBuildMerkleTreeFromLeaves_TwoLeaves(t *testing.T) {
	leaves := [][32]byte{
		decode32("0000000000000000000000000000000000000000000000000000000000000000"),
		decode32("1111111111111111111111111111111111111111111111111111111111111111"),
	}

	tree, err := BuildMerkleTreeFromLeaves(leaves)
	assert.NoError(t, err)

	// Root should be hash_pair of the two leaves
	expectedRoot := hashPair(leaves[0], leaves[1])
	assert.Equal(t, expectedRoot, tree.Root, "Two-leaf tree root should be hash_pair of leaves")

	// Both proofs should work
	for i, leaf := range leaves {
		valid := VerifyMerkleProof(leaf, tree.Proofs[i], tree.Root)
		assert.True(t, valid, "Proof for leaf %d must verify", i)
	}
}

// TestBuildMerkleTreeFromLeaves_OddNumber verifies handling of odd number of nodes
func TestBuildMerkleTreeFromLeaves_OddNumber(t *testing.T) {
	// Test with 5 leaves (odd at first level, then 3, then 2)
	leaves := [][32]byte{
		decode32("0000000000000000000000000000000000000000000000000000000000000000"),
		decode32("1111111111111111111111111111111111111111111111111111111111111111"),
		decode32("2222222222222222222222222222222222222222222222222222222222222222"),
		decode32("3333333333333333333333333333333333333333333333333333333333333333"),
		decode32("4444444444444444444444444444444444444444444444444444444444444444"),
	}

	tree, err := BuildMerkleTreeFromLeaves(leaves)
	assert.NoError(t, err)
	assert.NotEqual(t, [32]byte{}, tree.Root, "Root should not be empty")

	// All proofs should verify
	for i, leaf := range leaves {
		valid := VerifyMerkleProof(leaf, tree.Proofs[i], tree.Root)
		assert.True(t, valid, "Proof for leaf %d must verify", i)
	}
}

// TestBuildMerkleTreeFromLeaves_EmptyLeaves verifies error on empty input
func TestBuildMerkleTreeFromLeaves_EmptyLeaves(t *testing.T) {
	leaves := [][32]byte{}

	tree, err := BuildMerkleTreeFromLeaves(leaves)
	assert.Error(t, err, "Should error on empty leaves")
	assert.Nil(t, tree, "Tree should be nil on error")
	assert.Contains(t, err.Error(), "cannot build Merkle tree with no leaves")
}

// TestKeccak256Hash verifies keccak256 implementation
func TestKeccak256Hash(t *testing.T) {
	// Test empty input
	emptyHash := crypto.Keccak256Hash([]byte{})
	expected := decode32("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	assert.Equal(t, expected, [32]byte(emptyHash), "keccak256('') must match known value")

	// Test simple string
	helloHash := crypto.Keccak256Hash([]byte("hello"))
	expected = decode32("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	assert.Equal(t, expected, [32]byte(helloHash), "keccak256('hello') must match known value")
}
