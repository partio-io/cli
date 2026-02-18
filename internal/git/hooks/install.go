package hooks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jcleira/ai-workflow-core/internal/git"
)

// Install installs partio git hooks into the repository, backing up existing hooks.
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
		backupPath := hookPath + ".partio-backup"

		// If existing hook is not ours, back it up
		if data, err := os.ReadFile(hookPath); err == nil {
			content := string(data)
			if !isPartioHook(content) {
				if err := os.Rename(hookPath, backupPath); err != nil {
					return fmt.Errorf("backing up %s hook: %w", name, err)
				}
			}
		}

		// Write our hook
		script := hookScript(name)
		if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
			return fmt.Errorf("writing %s hook: %w", name, err)
		}
	}

	return nil
}
