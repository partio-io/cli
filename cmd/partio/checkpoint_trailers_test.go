package main

import (
	"reflect"
	"testing"
)

func TestCheckpointIDsFromMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     []string
	}{
		{
			name:     "several trailers in the final block",
			messages: []string{"feat: x\n\nBody.\n\nPartio-Checkpoint: aabbccddeeff\nPartio-Attribution: 100% agent\nPartio-Checkpoint: 112233445566\n"},
			want:     []string{"aabbccddeeff", "112233445566"},
		},
		{
			name:     "key matches case-insensitively",
			messages: []string{"fix: y\n\npartio-checkpoint: aabbccddeeff\n", "fix: z\n\nPARTIO-CHECKPOINT: 112233445566\n"},
			want:     []string{"aabbccddeeff", "112233445566"},
		},
		{
			name:     "a --- line in the body does not hide trailers",
			messages: []string{"docs: w\n\nSome text\n---\nmore markdown\n\nPartio-Checkpoint: aabbccddeeff\n"},
			want:     []string{"aabbccddeeff"},
		},
		{
			name:     "trailer-like line in an earlier paragraph is ignored",
			messages: []string{"feat: v\n\nPartio-Checkpoint: 000000000000\n\nA closing paragraph that is not a trailer block.\n"},
			want:     nil,
		},
		{
			name: "invalid values are skipped",
			messages: []string{
				"a\n\nPartio-Checkpoint: AABBCCDDEEFF\n",
				"b\n\nPartio-Checkpoint: abc\n",
				"c\n\nPartio-Checkpoint: ../../etc/pa\n",
				"d\n\nPartio-Checkpoint: abc\nPartio-Checkpoint: 112233445566\n",
			},
			want: []string{"112233445566"},
		},
		{
			name: "duplicates across commits keep first-seen order",
			messages: []string{
				"a\n\nPartio-Checkpoint: 112233445566\n",
				"b\n\nPartio-Checkpoint: aabbccddeeff\n",
				"c\n\nPartio-Checkpoint: 112233445566\n",
			},
			want: []string{"112233445566", "aabbccddeeff"},
		},
		{
			name:     "no trailers",
			messages: []string{"chore: nothing here\n"},
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkpointIDsFromMessages(t.TempDir(), tt.messages)
			if err != nil {
				t.Fatalf("checkpointIDsFromMessages: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
