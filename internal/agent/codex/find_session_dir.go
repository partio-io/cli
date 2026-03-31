package codex

import "fmt"

// FindSessionDir is not implemented for Codex CLI.
// Codex does not expose a stable per-project session directory.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	return "", fmt.Errorf("codex session directory not supported")
}
