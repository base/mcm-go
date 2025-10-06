// Package io provides JSON serialization/deserialization for proposal types.
package io

import (
	"encoding/hex"
	"fmt"

	hexutil "github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/proposal"

	"github.com/gagliardetto/solana-go"
)

// accountMetaJSON is the JSON representation of AccountMeta
type accountMetaJSON struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
}

// instructionJSON is the JSON representation of GenericInstruction
type instructionJSON struct {
	ProgramID string            `json:"programId"`
	Data      string            `json:"data"`
	Accounts  []accountMetaJSON `json:"accounts"`
}

// rootMetadataJSON is the JSON representation of RootMetadata
type rootMetadataJSON struct {
	ChainID              uint64 `json:"chainId"`
	Multisig             string `json:"multisig"`
	PreOpCount           uint64 `json:"preOpCount"`
	PostOpCount          uint64 `json:"postOpCount"`
	OverridePreviousRoot bool   `json:"overridePreviousRoot"`
}

// proposalJSON is the JSON representation of Proposal
type proposalJSON struct {
	MultisigID   string            `json:"multisigId"`
	ValidUntil   uint32            `json:"validUntil"`
	Instructions []instructionJSON `json:"instructions"`
	RootMetadata rootMetadataJSON  `json:"rootMetadata"`
}

// toProposalJSON converts a Proposal to its JSON DTO
func toProposalJSON(p *proposal.Proposal) (*proposalJSON, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal: %w", err)
	}

	ixsJSON := make([]instructionJSON, len(p.Instructions))
	for i, ix := range p.Instructions {
		accounts := make([]accountMetaJSON, len(ix.AccountValues))
		for j, acc := range ix.AccountValues {
			accounts[j] = accountMetaJSON{
				Pubkey:     acc.PublicKey.String(),
				IsSigner:   acc.IsSigner,
				IsWritable: acc.IsWritable,
			}
		}

		ixsJSON[i] = instructionJSON{
			ProgramID: ix.ProgID.String(),
			Data:      "0x" + hex.EncodeToString(ix.DataBytes),
			Accounts:  accounts,
		}
	}

	return &proposalJSON{
		MultisigID:   "0x" + hex.EncodeToString(p.MultisigID[:]),
		ValidUntil:   p.ValidUntil,
		Instructions: ixsJSON,
		RootMetadata: rootMetadataJSON{
			ChainID:              p.RootMetadata.ChainID,
			Multisig:             p.RootMetadata.Multisig.String(),
			PreOpCount:           p.RootMetadata.PreOpCount,
			PostOpCount:          p.RootMetadata.PostOpCount,
			OverridePreviousRoot: p.RootMetadata.OverridePreviousRoot,
		},
	}, nil
}

// fromProposalJSON converts a JSON DTO to a Proposal
func fromProposalJSON(pj *proposalJSON) (*proposal.Proposal, error) {
	multisigID, err := hexutil.Parse32(pj.MultisigID)
	if err != nil {
		return nil, fmt.Errorf("invalid multisigId: %w", err)
	}

	multisig, err := solana.PublicKeyFromBase58(pj.RootMetadata.Multisig)
	if err != nil {
		return nil, fmt.Errorf("invalid multisig pubkey: %w", err)
	}

	instructions := make([]*solana.GenericInstruction, len(pj.Instructions))
	for i, ixJSON := range pj.Instructions {
		progID, err := solana.PublicKeyFromBase58(ixJSON.ProgramID)
		if err != nil {
			return nil, fmt.Errorf("instruction %d: invalid programId: %w", i, err)
		}

		dataBytes, err := hexutil.Decode(ixJSON.Data)
		if err != nil {
			return nil, fmt.Errorf("instruction %d: invalid data hex: %w", i, err)
		}

		accounts := make([]*solana.AccountMeta, len(ixJSON.Accounts))
		for j, accJSON := range ixJSON.Accounts {
			pubkey, err := solana.PublicKeyFromBase58(accJSON.Pubkey)
			if err != nil {
				return nil, fmt.Errorf("instruction %d account %d: invalid pubkey: %w", i, j, err)
			}
			accounts[j] = &solana.AccountMeta{
				PublicKey:  pubkey,
				IsSigner:   accJSON.IsSigner,
				IsWritable: accJSON.IsWritable,
			}
		}

		instructions[i] = &solana.GenericInstruction{
			ProgID:        progID,
			DataBytes:     dataBytes,
			AccountValues: accounts,
		}
	}

	p := &proposal.Proposal{
		MultisigID:   multisigID,
		ValidUntil:   pj.ValidUntil,
		Instructions: instructions,
		RootMetadata: proposal.RootMetadata{
			ChainID:              pj.RootMetadata.ChainID,
			Multisig:             multisig,
			PreOpCount:           pj.RootMetadata.PreOpCount,
			PostOpCount:          pj.RootMetadata.PostOpCount,
			OverridePreviousRoot: pj.RootMetadata.OverridePreviousRoot,
		},
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal after load: %w", err)
	}

	return p, nil
}
