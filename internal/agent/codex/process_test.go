package codex

import (
	"fmt"
	"os/exec"
	"testing"
)

// makeExitError runs a subprocess with the given exit code to produce a real
// *exec.ExitError, which is required to satisfy the ExitCode() check.
func makeExitError(t *testing.T, code int) *exec.ExitError {
	t.Helper()
	cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", code))
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError from exit %d, got %T", code, err)
	}
	return exitErr
}

func TestParseIsRunning(t *testing.T) {
	tests := []struct {
		name    string
		out     []byte
		err     error
		want    bool
		wantErr bool
	}{
		{
			name: "output with PID returns true",
			out:  []byte("12345\n"),
			want: true,
		},
		{
			name: "multiple PIDs returns true",
			out:  []byte("12345\n67890\n"),
			want: true,
		},
		{
			name: "empty output returns false",
			out:  []byte(""),
			want: false,
		},
		{
			name: "whitespace-only output returns false",
			out:  []byte("   \n"),
			want: false,
		},
		{
			name: "pgrep exit code 1 (no match) returns false without error",
			err:  nil, // set dynamically below
			want: false,
		},
		{
			name:    "unexpected pgrep error returns error",
			err:     nil, // set dynamically below
			wantErr: true,
		},
	}

	// Fill in the dynamic exit errors.
	tests[4].err = makeExitError(t, 1)
	tests[5].err = makeExitError(t, 2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIsRunning(tt.out, tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIsRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseIsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
