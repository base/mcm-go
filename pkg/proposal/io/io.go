// Package io provides JSON persistence for proposal types.
package io

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/base/mcm-go/pkg/proposal"

	"github.com/gagliardetto/solana-go"
)

// SaveProposal saves a Proposal to a JSON file
func SaveProposal(p *proposal.Proposal, path string) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid proposal: %w", err)
	}

	dto, err := toProposalJSON(p)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadProposal loads a Proposal from a JSON file
func LoadProposal(path string) (*proposal.Proposal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dto proposalJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	p, err := fromProposalJSON(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from JSON: %w", err)
	}

	return p, nil
}

// instructionsOnlyJSON is a simplified JSON format containing only instructions
type instructionsOnlyJSON struct {
	Instructions []instructionJSON `json:"instructions"`
}

// LoadInstructions loads instructions from a simplified JSON file
// that contains only an array of instructions without metadata.
//
// Example JSON format:
//
//	{
//	  "instructions": [
//	    {
//	      "programId": "11111111111111111111111111111111",
//	      "data": "0xdeadbeef",
//	      "accounts": [
//	        {
//	          "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
//	          "isSigner": true,
//	          "isWritable": false
//	        }
//	      ]
//	    }
//	  ]
//	}
func LoadInstructions(path string) ([]*solana.GenericInstruction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dto instructionsOnlyJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(dto.Instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in file")
	}

	return parseInstructionsJSON(dto.Instructions)
}
