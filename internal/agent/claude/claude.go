package claude

import "strings"

// Detector implements the agent.Detector interface for Claude Code.
type Detector struct{}

// New creates a new Claude Code detector.
func New() *Detector {
	return &Detector{}
}

// Name returns the agent name.
func (d *Detector) Name() string {
	return "claude-code"
}

// sanitizePath converts an absolute path to Claude's sanitized format.
// e.g. /Users/foo/project -> -Users-foo-project
func sanitizePath(p string) string {
	return strings.ReplaceAll(p, "/", "-")
}
