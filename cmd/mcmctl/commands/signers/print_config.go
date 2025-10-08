package signers

import (
	"fmt"
	"sort"

	"github.com/base/mcm-go/cmd/mcmctl/flags"
	"github.com/base/mcm-go/cmd/mcmctl/util"
	"github.com/base/mcm-go/pkg/bindings"
	"github.com/base/mcm-go/pkg/hex"
	"github.com/base/mcm-go/pkg/state"

	ucli "github.com/urfave/cli/v2"
)

// PrintConfigCommand returns the multisig config command
func PrintConfigCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "print-config",
		Usage: "Display multisig configuration (signers, groups, quorums)",
		Flags: append(flags.OnchainReadFlags(),
			flags.MultisigIDFlag(),
			&ucli.BoolFlag{
				Name:  "pretty",
				Usage: "Display configuration as a tree hierarchy",
			},
		),
		Action: func(c *ucli.Context) error {
			// Load client
			mcmClient, err := util.LoadClient(c)
			if err != nil {
				return err
			}
			defer mcmClient.Close()

			multisigID, err := hex.Parse32(c.String("multisig-id"))
			if err != nil {
				return fmt.Errorf("invalid multisig-id: %w", err)
			}

			// Fetch MultisigConfig
			fetcher := state.NewFetcher(mcmClient.RPC, mcmClient.ProgramID)
			config, err := fetcher.GetMultisigConfig(c.Context, multisigID)
			if err != nil {
				return fmt.Errorf("failed to fetch multisig config: %w", err)
			}

			// Display configuration
			fmt.Println("=== Multisig Configuration ===")
			fmt.Printf("Multisig ID: 0x%x\n", config.MultisigId)
			fmt.Printf("Chain ID: %d\n", config.ChainId)
			fmt.Printf("Owner: %s\n", config.Owner)
			fmt.Printf("Proposed Owner: %s\n", config.ProposedOwner)

			if c.Bool("pretty") {
				printPrettyHierarchy(config)
			} else {
				printFlatConfig(config)
			}

			return nil
		},
	}
}

// printFlatConfig displays the configuration in a flat format
func printFlatConfig(config *bindings.MultisigConfig) {
	fmt.Printf("\n=== Signers (%d total) ===\n", len(config.Signers))
	for i, signer := range config.Signers {
		fmt.Printf("  [%d] Address: 0x%x, Index: %d, Group: %d\n",
			i, signer.EvmAddress, signer.Index, signer.Group)
	}

	fmt.Printf("\n=== Group Quorums ===\n")
	for i, quorum := range config.GroupQuorums {
		if quorum > 0 {
			fmt.Printf("  Group %d: quorum = %d\n", i, quorum)
		}
	}

	fmt.Printf("\n=== Group Parents (Hierarchy) ===\n")
	// Display groups that have a quorum set (indicating they exist)
	for i, parent := range config.GroupParents {
		if i < len(config.GroupQuorums) && config.GroupQuorums[i] > 0 {
			if i == 0 {
				fmt.Printf("  Group %d: ROOT (no parent)\n", i)
			} else {
				fmt.Printf("  Group %d: parent = Group %d\n", i, parent)
			}
		}
	}
}

// printPrettyHierarchy displays the configuration as a tree
func printPrettyHierarchy(config *bindings.MultisigConfig) {
	// Build group to children mapping
	groupChildren := make(map[uint8][]uint8)
	for groupID, parent := range config.GroupParents {
		if groupID == 0 {
			continue // Skip root (group 0 has no parent)
		}
		// Add this group as a child of its parent
		groupChildren[parent] = append(groupChildren[parent], uint8(groupID))
	}

	// Sort children for consistent display
	for _, children := range groupChildren {
		sort.Slice(children, func(i, j int) bool {
			return children[i] < children[j]
		})
	}

	// Build group to signers mapping
	groupSigners := make(map[uint8][]bindings.McmSigner)
	for _, signer := range config.Signers {
		groupSigners[signer.Group] = append(groupSigners[signer.Group], signer)
	}

	// Sort signers within each group by index
	for _, signers := range groupSigners {
		sort.Slice(signers, func(i, j int) bool {
			return signers[i].Index < signers[j].Index
		})
	}

	fmt.Println("\n=== Group Hierarchy ===")
	printGroup(0, "", true, config, groupChildren, groupSigners)
}

// printGroup recursively prints a group and its children in tree format
func printGroup(groupID uint8, prefix string, isLast bool, config *bindings.MultisigConfig, groupChildren map[uint8][]uint8, groupSigners map[uint8][]bindings.McmSigner) {
	// Check if group exists (has quorum > 0)
	quorum := uint8(0)
	if int(groupID) < len(config.GroupQuorums) {
		quorum = config.GroupQuorums[groupID]
	}
	if quorum == 0 {
		return // Skip groups that don't exist
	}

	// Print group header
	connector := "└── "
	if !isLast {
		connector = "├── "
	}

	if groupID == 0 {
		fmt.Printf("Group 0 (ROOT, quorum: %d)\n", quorum)
	} else {
		fmt.Printf("%s%sGroup %d (quorum: %d)\n", prefix, connector, groupID, quorum)
	}

	// Calculate prefix for children
	childPrefix := prefix
	if groupID != 0 {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	// Print signers in this group
	signers := groupSigners[groupID]

	// Filter children to only existing groups (those with quorum > 0)
	existingChildren := []uint8{}
	for _, childID := range groupChildren[groupID] {
		if int(childID) < len(config.GroupQuorums) && config.GroupQuorums[childID] > 0 {
			existingChildren = append(existingChildren, childID)
		}
	}

	for i, signer := range signers {
		isLastSigner := i == len(signers)-1 && len(existingChildren) == 0
		signerConnector := "└── "
		if !isLastSigner {
			signerConnector = "├── "
		}
		fmt.Printf("%s%s[Signer %d] 0x%x\n", childPrefix, signerConnector, signer.Index, signer.EvmAddress)
	}

	// Print child groups (only existing ones)
	for i, childID := range existingChildren {
		isLastChild := i == len(existingChildren)-1
		printGroup(childID, childPrefix, isLastChild, config, groupChildren, groupSigners)
	}
}
