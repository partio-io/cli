package hooks

import (
	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/agent/codex"
)

// resolveDetector returns the appropriate agent detector for the given name.
func resolveDetector(name string) agent.Detector {
	switch name {
	case "codex":
		return codex.New()
	default:
		return claude.New()
	}
}
