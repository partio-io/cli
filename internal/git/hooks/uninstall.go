package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// Uninstall removes the partio sentinel block from git hooks, leaving other content intact.
func Uninstall(repoRoot string) error {
	hooksDir, err := git.HooksDir(repoRoot)
	if err != nil {
		return err
	}

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksDir, name)

		data, err := os.ReadFile(hookPath)
		if err != nil {
			continue // Hook doesn't exist, skip
		}

		content := string(data)
		if !hasPartioBlock(content) {
			continue // No partio block, skip
		}

		stripped := strings.TrimSpace(removePartioBlock(content))

		if stripped == "" || stripped == "#!/bin/bash" {
			// Only partio content remained — remove the file
			_ = os.Remove(hookPath)
		} else {
			// Write back without the partio block
			if err := os.WriteFile(hookPath, []byte(stripped+"\n"), 0o755); err != nil {
				return fmt.Errorf("writing %s hook after removal: %w", name, err)
			}
		}
	}

	return nil
}
