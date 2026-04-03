package codex

import (
	"testing"

	"github.com/partio-io/cli/internal/agent"
)

func TestDetectorImplementsInterface(t *testing.T) {
	var _ agent.Detector = (*Detector)(nil)
}

func TestName(t *testing.T) {
	d := New()
	if got := d.Name(); got != "codex" {
		t.Errorf("Name() = %q, want %q", got, "codex")
	}
}
