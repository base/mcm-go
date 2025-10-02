// Package crypto provides cryptographic utilities for the MCM SDK,
// including Merkle tree construction and ECDSA signature handling.
package crypto

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/sha3"
)

// Proof represents a Merkle proof as a slice of 32-byte hashes
type Proof [][32]byte

// MerkleRootWithProofs contains the Merkle root and proofs for all leaves
type MerkleRootWithProofs struct {
	Root   [32]byte
	Proofs []Proof
}

// BuildMerkleTreeFromLeaves constructs a Merkle tree from the given leaves
// and returns the root hash along with proofs for each leaf.
// This implementation matches the MCM program's tree construction logic,
// which handles odd numbers of nodes by promoting the last node up.
func BuildMerkleTreeFromLeaves(leaves [][32]byte) (*MerkleRootWithProofs, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("cannot build Merkle tree with no leaves")
	}

	if len(leaves) == 1 {
		return &MerkleRootWithProofs{
			Root:   leaves[0],
			Proofs: []Proof{{}}, // Empty proof for single leaf
		}, nil
	}

	// Build the full tree dynamically
	tree := make([][][32]byte, 0)
	tree = append(tree, leaves)
	currentLevel := make([][32]byte, len(leaves))
	copy(currentLevel, leaves)

	// Build tree bottom-up, handling odd numbers like MCM
	for len(currentLevel) > 1 {
		nextLevel := make([][32]byte, 0)

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// Pair exists - hash both
				left := currentLevel[i]
				right := currentLevel[i+1]
				nextLevel = append(nextLevel, hashPair(left, right))
			} else {
				// Odd number - promote the last element up
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}

		tree = append(tree, nextLevel)
		currentLevel = nextLevel
	}

	// Generate proofs for each leaf
	proofs := make([]Proof, len(leaves))
	for i := range leaves {
		proofs[i] = generateMerkleProof(tree, i)
	}

	return &MerkleRootWithProofs{
		Root:   currentLevel[0],
		Proofs: proofs,
	}, nil
}

// hashPair hashes two 32-byte values using Keccak256.
// The hashes are sorted before hashing to ensure deterministic results.
func hashPair(a, b [32]byte) [32]byte {
	var left, right [32]byte

	// Sort the hashes lexicographically
	if bytes.Compare(a[:], b[:]) < 0 {
		left, right = a, b
	} else {
		left, right = b, a
	}

	// Concatenate and hash
	data := append(left[:], right[:]...)
	return Keccak256Hash(data)
}

// generateMerkleProof generates a Merkle proof for a leaf at the given index
func generateMerkleProof(tree [][][32]byte, leafIndex int) Proof {
	proof := make(Proof, 0)
	currentIndex := leafIndex

	// Traverse up the tree to collect siblings
	for level := 0; level < len(tree)-1; level++ {
		currentLevelNodes := tree[level]
		var siblingIndex int

		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}

		// Add sibling to proof if it exists
		if siblingIndex < len(currentLevelNodes) {
			proof = append(proof, currentLevelNodes[siblingIndex])
		}

		// Move to parent index in next level
		currentIndex = currentIndex / 2
	}

	return proof
}

// Keccak256Hash computes the Keccak256 hash of the input data
func Keccak256Hash(data []byte) [32]byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

// VerifyMerkleProof verifies that a leaf is part of a Merkle tree with the given root
func VerifyMerkleProof(leaf [32]byte, proof Proof, root [32]byte) bool {
	current := leaf

	for _, sibling := range proof {
		current = hashPair(current, sibling)
	}

	return bytes.Equal(current[:], root[:])
}
