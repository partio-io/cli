package agent

import "testing"

func TestNewDetector(t *testing.T) {
	// Register a test detector.
	Register("test-agent", func() Detector {
		return &stubDetector{name: "test-agent"}
	})

	tests := []struct {
		name    string
		agent   string
		want    string
		wantErr bool
	}{
		{
			name:  "registered agent",
			agent: "test-agent",
			want:  "test-agent",
		},
		{
			name:    "unknown agent",
			agent:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDetector(tt.agent)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDetector(%q) error = %v, wantErr %v", tt.agent, err, tt.wantErr)
				return
			}
			if err == nil && d.Name() != tt.want {
				t.Errorf("NewDetector(%q).Name() = %q, want %q", tt.agent, d.Name(), tt.want)
			}
		})
	}
}

type stubDetector struct {
	name string
}

func (s *stubDetector) Name() string                            { return s.name }
func (s *stubDetector) IsRunning() (bool, error)                { return false, nil }
func (s *stubDetector) FindSessionDir(repoRoot string) (string, error) {
	return "", nil
}
