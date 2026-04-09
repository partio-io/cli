package codex

// Detector implements the agent.Detector interface for OpenAI Codex CLI.
type Detector struct{}

// New creates a new Codex CLI detector.
func New() *Detector {
	return &Detector{}
}

// Name returns the agent name.
func (d *Detector) Name() string {
	return "codex"
}
