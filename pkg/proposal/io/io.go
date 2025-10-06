// Package io provides JSON persistence for proposal types.
package io

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/base/mcm-go/pkg/proposal"
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
