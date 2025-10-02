package proposal

import (
	"encoding/binary"
	"fmt"

	"mcm-go/pkg/crypto"

	"github.com/gagliardetto/solana-go"
)

// Domain separators for Merkle tree leaf hashing
var (
	// domainSeparatorMetadata is the domain separator for metadata leaves
	domainSeparatorMetadata = [32]byte{
		0x47, 0xfd, 0xed, 0x70, 0x90, 0x1d, 0x27, 0x03, 0x83, 0x94, 0xdb, 0x90, 0x5a,
		0x72, 0x56, 0x3c, 0xad, 0x6f, 0x07, 0x58, 0x1d, 0xbc, 0xdd, 0x14, 0x72, 0xcc,
		0xd2, 0xf7, 0x42, 0xaf, 0x63, 0x60,
	}

	// domainSeparatorOp is the domain separator for operation leaves
	domainSeparatorOp = [32]byte{
		0xfb, 0x98, 0x81, 0x6f, 0xf3, 0xc5, 0x13, 0x8a, 0x68, 0xab, 0xfd, 0x40, 0xb8,
		0xd8, 0xfb, 0xc2, 0x29, 0x72, 0xfe, 0xa1, 0xdd, 0x89, 0x75, 0x73, 0x31, 0x32,
		0x7e, 0x6e, 0x0a, 0x94, 0x40, 0xb7,
	}
)

// WithRoot computes the Merkle root and proofs for this proposal
func (p *Proposal) WithRoot() (*ProposalWithRoot, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal: %w", err)
	}

	// Encode metadata as a leaf
	metadataLeaf, err := encodeMetadataLeaf(p.RootMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to encode metadata: %w", err)
	}

	// Encode each operation as a leaf
	operationLeaves := make([][32]byte, len(p.Instructions))
	for i, ix := range p.Instructions {
		leaf, err := encodeOperationLeaf(
			p.RootMetadata.ChainID,
			p.RootMetadata.Multisig,
			uint64(p.RootMetadata.PreOpCount+uint64(i)),
			ix.ProgID,
			ix.DataBytes,
			ix.AccountValues,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to encode operation %d: %w", i, err)
		}
		operationLeaves[i] = leaf
	}

	// Build Merkle tree with metadata + operations
	allLeaves := make([][32]byte, 1+len(operationLeaves))
	allLeaves[0] = metadataLeaf
	copy(allLeaves[1:], operationLeaves)

	merkleTree, err := crypto.BuildMerkleTreeFromLeaves(allLeaves)
	if err != nil {
		return nil, fmt.Errorf("failed to build Merkle tree: %w", err)
	}

	// Extract proofs
	metadataProof := merkleTree.Proofs[0]
	operationProofs := merkleTree.Proofs[1:]

	return &ProposalWithRoot{
		Proposal: p,
		ProposalRoot: ProposalRoot{
			Root:            merkleTree.Root,
			MetadataProof:   metadataProof,
			OperationProofs: operationProofs,
		},
	}, nil
}

// encodeMetadataLeaf encodes RootMetadata into a Merkle leaf hash
func encodeMetadataLeaf(metadata RootMetadata) ([32]byte, error) {
	// keccak256(domainSeparator || chainId(32) || multisig(32) || preOpCount(32) || postOpCount(32) || overridePreviousRoot(32))
	data := make([]byte, 0, 32+32+32+32+32+32)

	// Domain separator (32 bytes)
	data = append(data, domainSeparatorMetadata[:]...)

	// chainId (u64, little-endian, left-padded to 32 bytes)
	chainIDPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(chainIDPadded[24:], metadata.ChainID)
	data = append(data, chainIDPadded...)

	// multisig (32 bytes)
	data = append(data, metadata.Multisig[:]...)

	// preOpCount (u64, little-endian, left-padded to 32 bytes)
	preOpCountPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(preOpCountPadded[24:], metadata.PreOpCount)
	data = append(data, preOpCountPadded...)

	// postOpCount (u64, little-endian, left-padded to 32 bytes)
	postOpCountPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(postOpCountPadded[24:], metadata.PostOpCount)
	data = append(data, postOpCountPadded...)

	// overridePreviousRoot (bool, left-padded to 32 bytes)
	overridePadded := make([]byte, 32)
	if metadata.OverridePreviousRoot {
		overridePadded[31] = 1
	}
	data = append(data, overridePadded...)

	return crypto.Keccak256Hash(data), nil
}

// encodeOperationLeaf encodes an operation into a Merkle leaf hash
func encodeOperationLeaf(
	chainID uint64,
	multisig solana.PublicKey,
	nonce uint64,
	to solana.PublicKey,
	data []byte,
	remainingAccounts []*solana.AccountMeta,
) ([32]byte, error) {
	// keccak256(domainSeparator || chainId(32) || multisig(32) || nonce(32) || to(32) || dataLen(32) || data || remainingAccountsLen(32) || serializedAccounts)
	buffer := make([]byte, 0)

	// Domain separator (32 bytes)
	buffer = append(buffer, domainSeparatorOp[:]...)

	// chainId (u64, little-endian, left-padded to 32 bytes)
	chainIDPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(chainIDPadded[24:], chainID)
	buffer = append(buffer, chainIDPadded...)

	// multisig (32 bytes)
	buffer = append(buffer, multisig[:]...)

	// nonce (u64, little-endian, left-padded to 32 bytes)
	noncePadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(noncePadded[24:], nonce)
	buffer = append(buffer, noncePadded...)

	// to (32 bytes)
	buffer = append(buffer, to[:]...)

	// dataLen (u64, little-endian, left-padded to 32 bytes)
	dataLenPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(dataLenPadded[24:], uint64(len(data)))
	buffer = append(buffer, dataLenPadded...)

	// data (raw bytes)
	buffer = append(buffer, data...)

	// remainingAccountsLen (u64, little-endian, left-padded to 32 bytes)
	accountsLenPadded := make([]byte, 32)
	binary.LittleEndian.PutUint64(accountsLenPadded[24:], uint64(len(remainingAccounts)))
	buffer = append(buffer, accountsLenPadded...)

	// serializedAccounts
	serializedAccounts := serializeRemainingAccounts(remainingAccounts)
	buffer = append(buffer, serializedAccounts...)

	return crypto.Keccak256Hash(buffer), nil
}

// serializeRemainingAccounts serializes accounts
func serializeRemainingAccounts(accounts []*solana.AccountMeta) []byte {
	if len(accounts) == 0 {
		return []byte{}
	}

	buffer := make([]byte, 0, len(accounts)*33) // 32 bytes pubkey + 1 byte flags

	for _, account := range accounts {
		// Address (32 bytes)
		buffer = append(buffer, account.PublicKey[:]...)

		// Flags (1 byte): bit 1 = isSigner, bit 0 = isWritable
		var flags byte
		if account.IsSigner {
			flags |= 0x02 // bit 1 set
		}
		if account.IsWritable {
			flags |= 0x01 // bit 0 set
		}
		buffer = append(buffer, flags)
	}

	return buffer
}
