package claude

import "testing"

func TestParseTasklistOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "claude.exe present",
			input: `"claude.exe","1234","Console","1","10,000 K"`,
			want:  true,
		},
		{
			name:  "Claude.exe mixed case",
			input: `"Claude.exe","5678","Console","1","20,000 K"`,
			want:  true,
		},
		{
			name:  "no claude process",
			input: "\"notepad.exe\",\"100\",\"Console\",\"1\",\"5,000 K\"\n\"explorer.exe\",\"200\",\"Console\",\"1\",\"50,000 K\"",
			want:  false,
		},
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "malformed line with no commas",
			input: "notaCSVline",
			want:  false,
		},
		{
			name:  "claude not in image name field",
			input: `"notepad.exe","1234","claude session","1","5,000 K"`,
			want:  false,
		},
		{
			name:  "multiple lines, claude in second",
			input: "\"explorer.exe\",\"100\",\"Console\",\"1\",\"50,000 K\"\n\"claude.exe\",\"200\",\"Console\",\"1\",\"10,000 K\"",
			want:  true,
		},
		{
			name:  "windows line endings",
			input: "\"claude.exe\",\"1234\",\"Console\",\"1\",\"10,000 K\"\r\n",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTasklistOutput(tt.input)
			if got != tt.want {
				t.Errorf("parseTasklistOutput(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
