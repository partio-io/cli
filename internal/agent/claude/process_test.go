package claude

import (
	"testing"
)

func TestParseTasklistOutput(t *testing.T) {
	tests := []struct {
		name  string
		csv   string
		want  bool
	}{
		{
			name: "contains claude.exe row",
			csv: `"claude.exe","12345","Console","1","50,000 K"` + "\n",
			want: true,
		},
		{
			name: "contains Claude.exe with mixed case",
			csv: `"notepad.exe","111","Console","1","1,000 K"` + "\n" +
				`"Claude.exe","999","Console","1","30,000 K"` + "\n",
			want: true,
		},
		{
			name: "no claude process",
			csv: `"notepad.exe","111","Console","1","1,000 K"` + "\n" +
				`"explorer.exe","222","Console","1","20,000 K"` + "\n",
			want: false,
		},
		{
			name: "empty output",
			csv:  "",
			want: false,
		},
		{
			name: "claude in non-image field should not match",
			csv:  `"notepad.exe","111","claude","1","1,000 K"` + "\n",
			want: false,
		},
		{
			name: "claude-code variant",
			csv: `"claude-code.exe","5678","Console","1","60,000 K"` + "\n",
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseTasklistOutput(tc.csv)
			if got != tc.want {
				t.Errorf("parseTasklistOutput(%q) = %v, want %v", tc.csv, got, tc.want)
			}
		})
	}
}
