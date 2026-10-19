package claude

import "testing"

func TestParseTasklistOutput(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		want bool
	}{
		{
			name: "claude.exe present",
			csv:  `"claude.exe","1234","Console","1","50,000 K"`,
			want: true,
		},
		{
			name: "Claude Code with mixed case",
			csv:  `"Claude.exe","5678","Console","1","100,000 K"`,
			want: true,
		},
		{
			name: "no claude process",
			csv: `"notepad.exe","111","Console","1","10,000 K"
"chrome.exe","222","Console","1","200,000 K"`,
			want: false,
		},
		{
			name: "multiple processes including claude",
			csv: `"svchost.exe","100","Services","0","5,000 K"
"claude.exe","999","Console","1","80,000 K"
"explorer.exe","200","Console","1","40,000 K"`,
			want: true,
		},
		{
			name: "empty output",
			csv:  "",
			want: false,
		},
		{
			name: "claude in non-image-name field does not match",
			// PID field contains "claude" — should not match since we only check the image name.
			csv: `"notepad.exe","claude","Console","1","10,000 K"`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTasklistOutput(tt.csv)
			if got != tt.want {
				t.Errorf("parseTasklistOutput(%q) = %v, want %v", tt.csv, got, tt.want)
			}
		})
	}
}
