package main

import (
	"strings"
	"testing"
)

func TestParsePRRef(t *testing.T) {
	tests := []struct {
		in      string
		want    prRef
		wantErr bool
	}{
		{in: "https://github.com/acme/widgets/pull/42", want: prRef{"acme", "widgets", 42}},
		{in: "https://github.com/acme/widgets/pull/42/files", want: prRef{"acme", "widgets", 42}},
		{in: "https://github.com/acme/widgets/pull/42/commits/", want: prRef{"acme", "widgets", 42}},
		{in: "https://ghe.example.com/acme/my.repo/pull/7", want: prRef{"acme", "my.repo", 7}},
		{in: "acme/widgets#42", want: prRef{"acme", "widgets", 42}},
		{in: "  acme/widgets#42\n", want: prRef{"acme", "widgets", 42}},
		{in: "https://github.com/acme/widgets/pull", wantErr: true},
		{in: "https://github.com/acme/widgets/pull/abc", wantErr: true},
		{in: "https://github.com/acme/widgets/pull/0", wantErr: true},
		{in: "https://github.com/acme/widgets/issues/42", wantErr: true},
		{in: "github.com/acme/widgets/pull/42", wantErr: true},
		{in: "42", wantErr: true},
		{in: "#42", wantErr: true},
		{in: "acme/widgets#", wantErr: true},
		{in: "acme/widgets#x1", wantErr: true},
		{in: "acme#42", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parsePRRef(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parsePRRef(%q) = %+v, want error", tt.in, got)
				}
				for _, form := range []string{"https://github.com/<owner>/<repo>/pull/<n>", "<owner>/<repo>#<n>"} {
					if !strings.Contains(err.Error(), form) {
						t.Errorf("error %q does not show %q", err, form)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePRRef(%q): %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("parsePRRef(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}
