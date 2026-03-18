package hooks

import (
	"testing"
)

func TestComputeCarryForward(t *testing.T) {
	tests := []struct {
		name           string
		allAgentFiles  []string
		committedFiles []string
		wantPending    []string
	}{
		{
			name:           "all committed - no carry-forward",
			allAgentFiles:  []string{"a.go", "b.go", "c.go"},
			committedFiles: []string{"a.go", "b.go", "c.go"},
			wantPending:    nil,
		},
		{
			name:           "partial commit - c.go carried forward",
			allAgentFiles:  []string{"a.go", "b.go", "c.go"},
			committedFiles: []string{"a.go", "b.go"},
			wantPending:    []string{"c.go"},
		},
		{
			name:           "no agent files - nothing to carry forward",
			allAgentFiles:  nil,
			committedFiles: []string{"a.go"},
			wantPending:    nil,
		},
		{
			name:           "multiple uncommitted files carried forward",
			allAgentFiles:  []string{"a.go", "b.go", "c.go", "d.go"},
			committedFiles: []string{"a.go"},
			wantPending:    []string{"b.go", "c.go", "d.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCarryForward(tt.allAgentFiles, tt.committedFiles)
			if len(got) != len(tt.wantPending) {
				t.Errorf("computeCarryForward() = %v, want %v", got, tt.wantPending)
				return
			}
			want := make(map[string]bool, len(tt.wantPending))
			for _, f := range tt.wantPending {
				want[f] = true
			}
			for _, f := range got {
				if !want[f] {
					t.Errorf("computeCarryForward() unexpected file %q in result %v", f, got)
				}
			}
		})
	}
}

func TestCheckCarryForwardActivation(t *testing.T) {
	tests := []struct {
		name            string
		cf              *carryForwardState
		stagedFiles     []string
		wantActivate    bool
		wantSessionPath string
	}{
		{
			name:         "no carry-forward state",
			cf:           nil,
			stagedFiles:  []string{"a.go"},
			wantActivate: false,
		},
		{
			name: "staged file matches pending - activation",
			cf: &carryForwardState{
				SessionPath:  "/path/to/session",
				PendingFiles: []string{"c.go"},
			},
			stagedFiles:     []string{"c.go"},
			wantActivate:    true,
			wantSessionPath: "/path/to/session",
		},
		{
			name: "staged file does not match pending - no activation",
			cf: &carryForwardState{
				SessionPath:  "/path/to/session",
				PendingFiles: []string{"c.go"},
			},
			stagedFiles:  []string{"a.go"},
			wantActivate: false,
		},
		{
			name: "empty pending files - no activation",
			cf: &carryForwardState{
				SessionPath:  "/path/to/session",
				PendingFiles: nil,
			},
			stagedFiles:  []string{"a.go"},
			wantActivate: false,
		},
		{
			name: "partial overlap activates",
			cf: &carryForwardState{
				SessionPath:  "/path/to/session",
				PendingFiles: []string{"c.go", "d.go"},
			},
			stagedFiles:     []string{"b.go", "c.go"},
			wantActivate:    true,
			wantSessionPath: "/path/to/session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activate, sessionPath := checkCarryForwardActivation(tt.cf, tt.stagedFiles)
			if activate != tt.wantActivate {
				t.Errorf("checkCarryForwardActivation() activate = %v, want %v", activate, tt.wantActivate)
			}
			if sessionPath != tt.wantSessionPath {
				t.Errorf("checkCarryForwardActivation() sessionPath = %q, want %q", sessionPath, tt.wantSessionPath)
			}
		})
	}
}

func TestLoadSaveCarryForward(t *testing.T) {
	dir := t.TempDir()

	// Load from empty dir returns nil, no error.
	cf, err := loadCarryForward(dir)
	if err != nil {
		t.Fatalf("loadCarryForward on empty dir: %v", err)
	}
	if cf != nil {
		t.Error("expected nil carry-forward initially")
	}

	// Save and reload.
	original := &carryForwardState{
		SessionPath:  "/path/to/session",
		PendingFiles: []string{"c.go", "d.go"},
		Branch:       "main",
	}
	if err := saveCarryForward(dir, original); err != nil {
		t.Fatalf("saveCarryForward: %v", err)
	}

	loaded, err := loadCarryForward(dir)
	if err != nil {
		t.Fatalf("loadCarryForward after save: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected carry-forward state after save")
	}
	if loaded.SessionPath != original.SessionPath {
		t.Errorf("SessionPath = %q, want %q", loaded.SessionPath, original.SessionPath)
	}
	if loaded.Branch != original.Branch {
		t.Errorf("Branch = %q, want %q", loaded.Branch, original.Branch)
	}
	if len(loaded.PendingFiles) != len(original.PendingFiles) {
		t.Errorf("PendingFiles len = %d, want %d", len(loaded.PendingFiles), len(original.PendingFiles))
	}

	// Clear and verify gone.
	clearCarryForward(dir)
	cleared, err := loadCarryForward(dir)
	if err != nil {
		t.Fatalf("loadCarryForward after clear: %v", err)
	}
	if cleared != nil {
		t.Error("expected nil carry-forward after clear")
	}
}

func TestCarryForwardPartialCommitScenario(t *testing.T) {
	// Scenario: agent modifies A, B, C; user commits only A and B.
	// Expect: carry-forward has C pending.
	allAgentFiles := []string{"a.go", "b.go", "c.go"}
	committedFiles := []string{"a.go", "b.go"}

	pending := computeCarryForward(allAgentFiles, committedFiles)

	if len(pending) != 1 || pending[0] != "c.go" {
		t.Errorf("expected carry-forward=[c.go], got %v", pending)
	}

	// Scenario: next commit stages C; carry-forward should activate.
	cf := &carryForwardState{
		SessionPath:  "/sessions/abc",
		PendingFiles: pending,
		Branch:       "main",
	}

	activate, sessionPath := checkCarryForwardActivation(cf, []string{"c.go"})
	if !activate {
		t.Error("expected carry-forward to activate for second commit")
	}
	if sessionPath != "/sessions/abc" {
		t.Errorf("expected sessionPath=/sessions/abc, got %q", sessionPath)
	}

	// After committing C, carry-forward should be empty.
	remaining := computeCarryForward(pending, []string{"c.go"})
	if len(remaining) != 0 {
		t.Errorf("expected no remaining files after committing C, got %v", remaining)
	}
}

func TestCarryForwardSingleCommitNoResidue(t *testing.T) {
	// Scenario: agent modifies A, B, C and user commits all at once.
	// Expect: no carry-forward.
	allAgentFiles := []string{"a.go", "b.go", "c.go"}
	committedFiles := []string{"a.go", "b.go", "c.go"}

	pending := computeCarryForward(allAgentFiles, committedFiles)
	if len(pending) != 0 {
		t.Errorf("expected no carry-forward for single full commit, got %v", pending)
	}
}

func TestMergeFiles(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want []string
	}{
		{
			name: "no overlap",
			a:    []string{"a.go", "b.go"},
			b:    []string{"c.go"},
			want: []string{"a.go", "b.go", "c.go"},
		},
		{
			name: "with overlap - deduplicated",
			a:    []string{"a.go", "b.go"},
			b:    []string{"b.go", "c.go"},
			want: []string{"a.go", "b.go", "c.go"},
		},
		{
			name: "empty b",
			a:    []string{"a.go"},
			b:    nil,
			want: []string{"a.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeFiles(tt.a, tt.b)
			if len(got) != len(tt.want) {
				t.Errorf("mergeFiles() = %v, want %v", got, tt.want)
				return
			}
			wantSet := make(map[string]bool, len(tt.want))
			for _, f := range tt.want {
				wantSet[f] = true
			}
			for _, f := range got {
				if !wantSet[f] {
					t.Errorf("mergeFiles() unexpected file %q", f)
				}
			}
		})
	}
}
