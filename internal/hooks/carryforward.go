package hooks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const carryForwardFile = "carry-forward.json"

// carryForwardState records agent-modified files not yet committed, enabling
// agent attribution to carry over to subsequent partial commits.
type carryForwardState struct {
	SessionPath  string   `json:"session_path"`
	PendingFiles []string `json:"pending_files"`
	Branch       string   `json:"branch"`
}

func loadCarryForward(stateDir string) (*carryForwardState, error) {
	data, err := os.ReadFile(filepath.Join(stateDir, carryForwardFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cf carryForwardState
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

func saveCarryForward(stateDir string, cf *carryForwardState) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cf)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, carryForwardFile), data, 0o644)
}

func clearCarryForward(stateDir string) {
	_ = os.Remove(filepath.Join(stateDir, carryForwardFile))
}

// computeCarryForward returns files from allAgentFiles that were not committed.
func computeCarryForward(allAgentFiles, committedFiles []string) []string {
	if len(allAgentFiles) == 0 {
		return nil
	}
	committed := make(map[string]bool, len(committedFiles))
	for _, f := range committedFiles {
		committed[f] = true
	}
	var pending []string
	for _, f := range allAgentFiles {
		if !committed[f] {
			pending = append(pending, f)
		}
	}
	return pending
}

// checkCarryForwardActivation returns whether a carry-forward state should
// activate agent attribution for the given staged files, and the session path.
func checkCarryForwardActivation(cf *carryForwardState, stagedFiles []string) (activate bool, sessionPath string) {
	if cf == nil || len(cf.PendingFiles) == 0 {
		return false, ""
	}
	staged := make(map[string]bool, len(stagedFiles))
	for _, f := range stagedFiles {
		staged[f] = true
	}
	for _, f := range cf.PendingFiles {
		if staged[f] {
			return true, cf.SessionPath
		}
	}
	return false, ""
}

// mergeFiles returns the union of two file slices without duplicates.
func mergeFiles(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	var result []string
	for _, f := range a {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	for _, f := range b {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	return result
}
