package hooks

import (
	"testing"

	"github.com/partio-io/cli/internal/config"
)

func TestShouldLinkCommit(t *testing.T) {
	tests := []struct {
		name     string
		linking  string
		wantLink bool
	}{
		{
			name:     "always links without prompt",
			linking:  config.CommitLinkingAlways,
			wantLink: true,
		},
		{
			name:     "never skips without prompt",
			linking:  config.CommitLinkingNever,
			wantLink: false,
		},
		{
			name:     "ask defaults to link when no TTY",
			linking:  config.CommitLinkingAsk,
			wantLink: true, // no TTY in test environment → auto-link
		},
		{
			name:     "empty defaults to link when no TTY",
			linking:  "",
			wantLink: true, // treated as "ask" → no TTY → auto-link
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{CommitLinking: tt.linking}
			got := shouldLinkCommit(t.TempDir(), cfg)
			if got != tt.wantLink {
				t.Errorf("shouldLinkCommit() = %v, want %v", got, tt.wantLink)
			}
		})
	}
}
