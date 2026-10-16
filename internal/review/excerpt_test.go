package review

import "testing"

func TestExcerptDropsOnlyHeadingsAndComments(t *testing.T) {
	tests := []struct {
		name, body, want string
	}{
		{"heading", "## Description\n\nFix it.", "Fix it."},
		{"issue number at line start", "## Why\n\n#771 is the same idea.", "#771 is the same idea."},
		{"comment", "## Premise\n\n<!-- partio:premise:v1 -->\n\n- A claim.", "- A claim."},
		{"crlf heading", "## Summary\r\n\r\nGoose fires hooks.", "Goose fires hooks."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := excerpt(tt.body); got != tt.want {
				t.Errorf("excerpt(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}
