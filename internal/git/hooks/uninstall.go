package hooks

import (
	"os"
	"path/filepath"

	"github.com/jcleira/ai-workflow-core/internal/git"
)

// Uninstall removes partio git hooks, restoring backups if present.
func Uninstall(repoRoot string) error {
	hooksDir, err := git.HooksDir(repoRoot)
	if err != nil {
		return err
	}

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksDir, name)
		backupPath := hookPath + ".partio-backup"

		// Only remove if it's our hook
		if data, err := os.ReadFile(hookPath); err == nil {
			if isPartioHook(string(data)) {
				_ = os.Remove(hookPath)
			}
		}

		// Restore backup if present
		if _, err := os.Stat(backupPath); err == nil {
			_ = os.Rename(backupPath, hookPath)
		}
	}

	return nil
}
