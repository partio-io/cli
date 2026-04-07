package claude

import "github.com/partio-io/cli/internal/agent"

func init() {
	agent.Register("claude-code", func() agent.Detector { return New() })
}
