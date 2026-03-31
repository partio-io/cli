package codex

import (
	"os/exec"
	"testing"
)

func TestParseIsRunning(t *testing.T) {
	// Obtain a real *exec.ExitError with exit code 1 for the "not found" test case.
	// pgrep exits with code 1 when no processes match; we replicate that here.
	var exitErr1 error
	cmd := exec.Command("sh", "-c", "exit 1")
	if err := cmd.Run(); err != nil {
		exitErr1 = err
	}

	tests := []struct {
		name    string
		out     []byte
		err     error
		want    bool
		wantErr bool
	}{
		{
			name: "empty output returns false",
			out:  []byte(""),
			err:  nil,
			want: false,
		},
		{
			name: "whitespace-only output returns false",
			out:  []byte("   \n"),
			err:  nil,
			want: false,
		},
		{
			name: "pid in output returns true",
			out:  []byte("12345\n"),
			err:  nil,
			want: true,
		},
		{
			name: "multiple pids returns true",
			out:  []byte("12345\n67890\n"),
			err:  nil,
			want: true,
		},
		{
			name: "pgrep exit code 1 means not found — not an error",
			out:  []byte(""),
			err:  exitErr1,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIsRunning(tt.out, tt.err)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseIsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectorName(t *testing.T) {
	d := New()
	if got := d.Name(); got != "codex" {
		t.Errorf("Name() = %q, want %q", got, "codex")
	}
}
