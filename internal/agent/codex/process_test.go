package codex

import (
	"os/exec"
	"testing"
)

func TestIsRunning(t *testing.T) {
	// Generate a real exit-1 error to simulate pgrep finding no processes.
	_, pgrepNotFoundErr := exec.Command("bash", "-c", "exit 1").Output()

	tests := []struct {
		name    string
		out     []byte
		err     error
		want    bool
		wantErr bool
	}{
		{
			name: "process running",
			out:  []byte("1234\n"),
			err:  nil,
			want: true,
		},
		{
			name: "process running multiple pids",
			out:  []byte("1234\n5678\n"),
			err:  nil,
			want: true,
		},
		{
			name: "process not running returns exit 1",
			out:  nil,
			err:  pgrepNotFoundErr,
			want: false,
		},
		{
			name: "empty output means not running",
			out:  []byte(""),
			err:  nil,
			want: false,
		},
		{
			name: "whitespace only output means not running",
			out:  []byte("   \n"),
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := execCommand
			defer func() { execCommand = orig }()

			execCommand = func(name string, args ...string) ([]byte, error) {
				return tt.out, tt.err
			}

			d := New()
			got, err := d.IsRunning()
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
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
