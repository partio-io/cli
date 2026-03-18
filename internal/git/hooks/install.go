package hooks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/git"
)

// Install installs partio git hooks into the repository, appending to existing hooks.
func Install(repoRoot string) error {
	hooksDir, err := git.HooksDir(repoRoot)
	if err != nil {
		return fmt.Errorf("resolving hooks directory: %w", err)
	}

	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksDir, name)

		data, err := os.ReadFile(hookPath)
		if err != nil {
			// No existing hook — create a new one
			script := newHookScript(name)
			if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
				return fmt.Errorf("writing %s hook: %w", name, err)
			}
			continue
		}

		if hasPartioBlock(string(data)) {
			// Already installed, skip
			continue
		}

		// Append partio block to existing hook
		block := "\n" + partioBlock(name) + "\n"
		f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0o755)
		if err != nil {
			return fmt.Errorf("opening %s hook for append: %w", name, err)
		}
		_, writeErr := f.WriteString(block)
		closeErr := f.Close()
		if writeErr != nil {
			return fmt.Errorf("appending to %s hook: %w", name, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("closing %s hook: %w", name, closeErr)
		}
	}

	return nil
}
