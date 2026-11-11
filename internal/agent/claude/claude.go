package claude

import "regexp"

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

// nonAlnumRe matches any character that is not an ASCII letter or digit.
var nonAlnumRe = regexp.MustCompile(`[^a-zA-Z0-9]`)

// sanitizePath converts an absolute path to Claude's sanitized format by
// replacing every non-alphanumeric character with "-".
// e.g. /Users/foo/project -> -Users-foo-project
// e.g. C:\Users\user\src  -> C--Users-user-src
func sanitizePath(p string) string {
	return nonAlnumRe.ReplaceAllString(p, "-")
}
