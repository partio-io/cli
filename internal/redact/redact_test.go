package redact

import (
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/checkpoint"
)

func TestText_Disabled(t *testing.T) {
	opts := DefaultOptions()
	opts.Enabled = false
	secret := "AKIAIOSFODNN7EXAMPLE"
	if got := Text(secret, opts); got != secret {
		t.Errorf("expected input unchanged when disabled, got %q", got)
	}
}

func TestText_Empty(t *testing.T) {
	opts := DefaultOptions()
	if got := Text("", opts); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// True-positive: known secret formats must be redacted.
var truePositives = []struct {
	name  string
	input string
}{
	{"aws-access-key-id", "AKIAIOSFODNN7EXAMPLE"},
	{"aws-asia-key", "ASIAIOSFODNN7EXAMPLE1"},
	{"github-pat", "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
	{"github-oauth", "gho_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
	{"github-app-token-ghu", "ghu_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
	{"github-app-token-ghs", "ghs_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
	{"github-fine-grained", "github_pat_" + strings.Repeat("A", 82)},
	{"slack-token", "xoxb-1234567890-abcdefghijklmnop"},
	{"stripe-live-key", "sk_live_abcdefghijklmnopqrstuvwx"},
	{"stripe-test-key", "sk_test_abcdefghijklmnopqrstuvwx"},
	{"google-api-key", "AIzaSyDabcdefghijklmnopqrstuvwxyz1234567"},
	{"npm-token", "npm_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
	{"pypi-token", "pypi-" + strings.Repeat("A", 50)},
	{"generic-api-key-assignment", "api_key=supersecretvaluehere1234567"},
	{"bearer-token", "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"},
	{"private-key-header", "-----BEGIN RSA PRIVATE KEY-----"},
}

func TestText_TruePositives(t *testing.T) {
	opts := DefaultOptions()
	for _, tc := range truePositives {
		t.Run(tc.name, func(t *testing.T) {
			got := Text(tc.input, opts)
			if !strings.Contains(got, Redacted) {
				t.Errorf("expected [REDACTED] in output for %q\ngot: %q", tc.input, got)
			}
		})
	}
}

// False-positives: these should NOT be redacted.
var falsePositives = []struct {
	name  string
	input string
}{
	{"plain-english", "The quick brown fox jumps over the lazy dog"},
	{"short-word", "hello"},
	{"commit-hash", "a3f1b2c4d5e6"},         // short, low-entropy
	{"version-string", "v1.2.3-alpha"},       // short
	{"url", "https://example.com/api/v1"},    // URL
	{"local-path", "/usr/local/bin/partio"},  // file path
	{"relative-path", "./internal/redact/"},  // relative path
	{"go-import", "github.com/partio-io/cli"}, // looks like a path
	{"markdown-header", "# Secret Detection"},
	// [REDACTED] is intentionally excluded from this table because it always
	// contains the placeholder — see TestText_AlreadyRedactedUnchanged.
	{"log-line", "2024-01-01T00:00:00Z INFO checkpoint created id=abc123"},
}

func TestText_FalsePositives(t *testing.T) {
	opts := DefaultOptions()
	for _, tc := range falsePositives {
		t.Run(tc.name, func(t *testing.T) {
			got := Text(tc.input, opts)
			if strings.Contains(got, Redacted) {
				t.Errorf("unexpected [REDACTED] for false-positive %q\ngot: %q", tc.input, got)
			}
		})
	}
}

// TestText_AlreadyRedactedUnchanged ensures the placeholder is not double-redacted.
func TestText_AlreadyRedactedUnchanged(t *testing.T) {
	opts := DefaultOptions()
	input := "[REDACTED]"
	got := Text(input, opts)
	if got != input {
		t.Errorf("expected [REDACTED] to pass through unchanged, got %q", got)
	}
}

func TestText_PatternInContext(t *testing.T) {
	opts := DefaultOptions()
	input := `config file:
api_key=supersecretvaluehere1234567
other_setting=normal_value`
	got := Text(input, opts)
	if !strings.Contains(got, Redacted) {
		t.Errorf("expected [REDACTED] in output\ngot: %q", got)
	}
	if !strings.Contains(got, "other_setting=normal_value") {
		t.Errorf("expected non-secret lines preserved\ngot: %q", got)
	}
}

func TestEntropyRedaction(t *testing.T) {
	opts := DefaultOptions()
	// A high-entropy base64-like string that doesn't match any known pattern.
	highEntropy := "xK9mP2qR7sT4uV1wY6zA3bC8dE5fG0hI"
	got := Text(highEntropy, opts)
	if !strings.Contains(got, Redacted) {
		t.Errorf("expected high-entropy string to be redacted\ngot: %q", got)
	}
}

func TestEntropyThreshold(t *testing.T) {
	opts := DefaultOptions()
	opts.EntropyThreshold = 10.0 // impossibly high — nothing should be flagged
	// A high-entropy base64-like string.
	highEntropy := "xK9mP2qR7sT4uV1wY6zA3bC8dE5fG0hI"
	got := Text(highEntropy, opts)
	if strings.Contains(got, Redacted) {
		t.Errorf("expected no redaction with very high threshold, got %q", got)
	}
}

func TestSessionFiles_Redaction(t *testing.T) {
	opts := DefaultOptions()
	sf := &checkpoint.SessionFiles{
		ContentHash: "deadbeefdeadbeef", // must NOT be redacted
		Prompt:      "my api_key=supersecretvaluehere1234567 please help",
		Context:     "user context",
		Diff:        "diff content with AKIAIOSFODNN7EXAMPLE inside",
		FullJSONL:   `{"role":"user","content":"token: ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"}`,
		Plan:        "normal plan text",
	}

	result := SessionFiles(sf, opts)

	if strings.Contains(result.Prompt, "supersecretvaluehere1234567") {
		t.Errorf("prompt secret not redacted: %q", result.Prompt)
	}
	if strings.Contains(result.Diff, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("diff secret not redacted: %q", result.Diff)
	}
	if strings.Contains(result.FullJSONL, "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij") {
		t.Errorf("full_jsonl secret not redacted: %q", result.FullJSONL)
	}
	// Content hash must be preserved.
	if result.ContentHash != "deadbeefdeadbeef" {
		t.Errorf("content hash must not be redacted, got %q", result.ContentHash)
	}
	// Normal plan text must be preserved.
	if result.Plan != "normal plan text" {
		t.Errorf("plan was unexpectedly modified: %q", result.Plan)
	}
}

func TestSessionFiles_Nil(t *testing.T) {
	opts := DefaultOptions()
	if got := SessionFiles(nil, opts); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSessionFiles_Disabled(t *testing.T) {
	opts := DefaultOptions()
	opts.Enabled = false
	sf := &checkpoint.SessionFiles{Prompt: "AKIAIOSFODNN7EXAMPLE"}
	result := SessionFiles(sf, opts)
	if result.Prompt != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected prompt unchanged when disabled, got %q", result.Prompt)
	}
}

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		s    string
		want float64 // approximate
		desc string
	}{
		{"", 0, "empty string"},
		{"aaaa", 0, "all same chars"},
		{"ab", 1, "two distinct chars equally split"},
	}
	for _, tc := range tests {
		got := shannonEntropy(tc.s)
		if got != tc.want {
			t.Errorf("shannonEntropy(%q) = %f, want %f (%s)", tc.s, got, tc.want, tc.desc)
		}
	}

	// High-entropy string should score above 4.0.
	h := shannonEntropy("xK9mP2qR7sT4uV1wY6zA3bC8dE5fG0hI")
	if h < 4.0 {
		t.Errorf("expected high entropy for random-looking string, got %f", h)
	}
}
