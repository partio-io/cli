package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	// Write test JSONL data
	lines := `{"type":"human","role":"human","message":"Build a todo app","timestamp":1700000000,"sessionId":"sess-123"}
{"type":"assistant","role":"assistant","message":"I'll help you build a todo app.","timestamp":1700000010}
{"type":"human","role":"human","message":"Add a delete button","timestamp":1700000020}
{"type":"assistant","role":"assistant","message":"Done, I added the delete button.","timestamp":1700000030}
`

	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	data, err := ParseJSONL(path)
	if err != nil {
		t.Fatalf("ParseJSONL error: %v", err)
	}

	if data.SessionID != "sess-123" {
		t.Errorf("expected session ID sess-123, got %s", data.SessionID)
	}

	if data.Agent != "claude-code" {
		t.Errorf("expected agent claude-code, got %s", data.Agent)
	}

	if len(data.Transcript) != 4 {
		t.Errorf("expected 4 messages, got %d", len(data.Transcript))
	}

	if data.Prompt != "Build a todo app" {
		t.Errorf("expected prompt 'Build a todo app', got %q", data.Prompt)
	}
}

func TestParseJSONLWithContentBlocks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	lines := `{"type":"human","role":"human","contentBlocks":[{"type":"text","text":"Hello world"}],"timestamp":1700000000}
{"type":"assistant","role":"assistant","contentBlocks":[{"type":"text","text":"Hi there!"}],"timestamp":1700000010}
`

	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	data, err := ParseJSONL(path)
	if err != nil {
		t.Fatalf("ParseJSONL error: %v", err)
	}

	if len(data.Transcript) != 2 {
		t.Errorf("expected 2 messages, got %d", len(data.Transcript))
	}

	if data.Transcript[0].Content != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", data.Transcript[0].Content)
	}
}

func TestParseJSONLEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	data, err := ParseJSONL(path)
	if err != nil {
		t.Fatalf("ParseJSONL error: %v", err)
	}

	if len(data.Transcript) != 0 {
		t.Errorf("expected 0 messages, got %d", len(data.Transcript))
	}
}

func TestParseJSONLWithSlug(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	lines := `{"type":"human","role":"human","message":"Build a todo app","timestamp":1700000000,"sessionId":"sess-456","slug":"noble-mixing-unicorn"}
{"type":"assistant","role":"assistant","message":"I'll help you build a todo app.","timestamp":1700000010}
`

	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	data, err := ParseJSONL(path)
	if err != nil {
		t.Fatalf("ParseJSONL error: %v", err)
	}

	if data.PlanSlug != "noble-mixing-unicorn" {
		t.Errorf("expected plan slug 'noble-mixing-unicorn', got %q", data.PlanSlug)
	}
}

func TestParseJSONLWithoutSlug(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	lines := `{"type":"human","role":"human","message":"Hello","timestamp":1700000000,"sessionId":"sess-789"}
`

	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	data, err := ParseJSONL(path)
	if err != nil {
		t.Fatalf("ParseJSONL error: %v", err)
	}

	if data.PlanSlug != "" {
		t.Errorf("expected empty plan slug, got %q", data.PlanSlug)
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/foo/project", "-Users-foo-project"},
		{"/home/user/code", "-home-user-code"},
	}

	for _, tt := range tests {
		if got := sanitizePath(tt.input); got != tt.expected {
			t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
