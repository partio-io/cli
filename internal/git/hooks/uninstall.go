package hooks

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/git"
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
				if rmErr := os.Remove(hookPath); rmErr != nil {
					slog.Warn("could not remove partio hook", "path", hookPath, "error", rmErr)
				}
			}
		}

		// Restore backup if present
		if _, err := os.Stat(backupPath); err == nil {
			if renameErr := os.Rename(backupPath, hookPath); renameErr != nil {
				slog.Warn("could not restore hook backup", "backup", backupPath, "path", hookPath, "error", renameErr)
			}
		}
	}

	return nil
}
