package attribution

import (
	"strconv"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// Calculate computes attribution for a commit based on whether an agent was active.
func Calculate(commitHash string, agentActive bool) (*Result, error) {
	// DiffNumstat uses git diff-tree --root, which handles initial commits natively.
	numstat, err := git.DiffNumstat(commitHash)
	if err != nil {
		return &Result{}, nil
	}

	totalAdded := 0
	for _, line := range strings.Split(numstat, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		added, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // binary file
		}
		totalAdded += added
	}

	result := &Result{
		TotalLines: totalAdded,
	}

	if agentActive {
		result.AgentLines = totalAdded
		result.AgentPercent = 100
	} else {
		result.HumanLines = totalAdded
		result.AgentPercent = 0
	}

	return result, nil
}
