package codex

import "strings"

// Detector implements the agent.Detector interface for Codex CLI.
type Detector struct{}

// New creates a new Codex CLI detector.
func New() *Detector {
	return &Detector{}
}

// Name returns the agent name.
func (d *Detector) Name() string {
	return "codex"
}

// sanitizePath converts an absolute path to a sanitized directory name.
// e.g. /Users/foo/project -> -Users-foo-project
func sanitizePath(p string) string {
	return strings.ReplaceAll(p, "/", "-")
}
