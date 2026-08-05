package main

import (
	"testing"
	"time"

	"github.com/partio-io/cli/internal/checkpoint"
)

func TestPickMostRecentCheckpoint(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name        string
		checkpoints []checkpoint.Checkpoint
		wantID      string
		wantErr     bool
	}{
		{
			name:        "empty slice returns error",
			checkpoints: []checkpoint.Checkpoint{},
			wantErr:     true,
		},
		{
			name: "single checkpoint returned",
			checkpoints: []checkpoint.Checkpoint{
				{ID: "aabbcc001122", CreatedAt: now},
			},
			wantID: "aabbcc001122",
		},
		{
			name: "multiple checkpoints returns most recent",
			checkpoints: []checkpoint.Checkpoint{
				{ID: "older00000001", CreatedAt: now.Add(-2 * time.Hour)},
				{ID: "newest0000002", CreatedAt: now},
				{ID: "middle0000003", CreatedAt: now.Add(-1 * time.Hour)},
			},
			wantID: "newest0000002",
		},
		{
			name: "tie broken by stable sort preserving order",
			checkpoints: []checkpoint.Checkpoint{
				{ID: "first00000001", CreatedAt: now},
				{ID: "second0000002", CreatedAt: now},
			},
			wantID: "first00000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pickMostRecentCheckpoint(tt.checkpoints)
			if (err != nil) != tt.wantErr {
				t.Fatalf("pickMostRecentCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("pickMostRecentCheckpoint() = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}

func TestResumeCommandArgsValidation(t *testing.T) {
	cmd := newResumeCmd()

	// No arguments and no --branch-name should fail validation.
	cmd.SetArgs([]string{})
	err := cmd.ValidateArgs([]string{})
	if err != nil {
		// cobra.MaximumNArgs(1) with 0 args is valid at the cobra level;
		// the business-logic check is in RunE, not ValidateArgs.
		// So this should not error at the cobra validation level.
		t.Errorf("unexpected cobra validation error with 0 args: %v", err)
	}

	// Two positional arguments should fail cobra validation.
	err = cmd.ValidateArgs([]string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected cobra validation error for 2 positional args, got nil")
	}
}
