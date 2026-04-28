package codex

import "github.com/partio-io/cli/internal/agent"

func init() {
	agent.Register("codex", func() agent.Detector { return New() })
}
