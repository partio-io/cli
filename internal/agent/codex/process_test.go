package codex

import (
	"strings"
	"testing"
)

func TestParsePgrepOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "single PID",
			output: "12345\n",
			want:   true,
		},
		{
			name:   "multiple PIDs",
			output: "12345\n67890\n",
			want:   true,
		},
		{
			name:   "empty output",
			output: "",
			want:   false,
		},
		{
			name:   "whitespace only",
			output: "  \n  ",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.TrimSpace(tt.output) != ""
			if got != tt.want {
				t.Errorf("parsePgrepOutput(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}
