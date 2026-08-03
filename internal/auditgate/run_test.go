package auditgate

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fakeGitHub is a minimal issues-comments API double. It serves the
// comments it is seeded with and records every comment created or
// updated through it.
type fakeGitHub struct {
	comments []map[string]any
	created  []string
	updated  map[int64]string
}

func newFakeGitHub(seed ...map[string]any) *fakeGitHub {
	return &fakeGitHub{comments: seed, updated: map[int64]string{}}
}

func (f *fakeGitHub) server(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/issues/41/comments", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(f.comments); err != nil {
			t.Errorf("encode comments: %v", err)
		}
	})
	mux.HandleFunc("POST /repos/partio-io/cli/issues/41/comments", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode created comment: %v", err)
		}
		f.created = append(f.created, payload.Body)
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id":1}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	mux.HandleFunc("PATCH /repos/partio-io/cli/issues/comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode updated comment: %v", err)
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			t.Errorf("parse comment id: %v", err)
		}
		f.updated[id] = payload.Body
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeVerdict(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write verdict fixture: %v", err)
	}
	return path
}

func runGate(t *testing.T, gh *fakeGitHub, verdictPath string) Result {
	t.Helper()
	srv := gh.server(t)
	res, err := Run(Config{
		VerdictPath: verdictPath,
		Repo:        "partio-io/cli",
		PRNumber:    41,
		AuditName:   "dead-code",
		APIBaseURL:  srv.URL,
		Token:       "test-token",
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res
}

// A pass verdict is the only green outcome, and it must leave the PR
// untouched: no comment created, none updated.
func TestRunPassVerdictStaysGreenAndSilent(t *testing.T) {
	verdict := writeVerdict(t, `{"status": "pass", "findings": []}`)
	gh := newFakeGitHub()

	res := runGate(t, gh, verdict)

	if !res.Green {
		t.Fatalf("Run returned red for a pass verdict: %+v", res)
	}
	if len(gh.created) != 0 || len(gh.updated) != 0 {
		t.Errorf("pass verdict touched comments: created=%q updated=%v", gh.created, gh.updated)
	}
}

// A missing verdict file means the audit session never finished its
// final act. The gate fails closed: red, with a comment saying the
// verdict was unusable rather than silently swallowing the outage.
func TestRunMissingVerdictFailsClosed(t *testing.T) {
	gh := newFakeGitHub()

	res := runGate(t, gh, filepath.Join(t.TempDir(), "never-written.json"))

	if res.Green {
		t.Fatalf("Run returned green for a missing verdict file")
	}
	if len(gh.created) != 1 {
		t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	if !strings.HasPrefix(gh.created[0], "Minion audit — dead-code") {
		t.Errorf("fail-closed comment does not start with audit prefix: %q", gh.created[0])
	}
}

// Malformed verdicts are as untrustworthy as missing ones: red, with
// the fail-closed comment, never a green or a findings-style comment
// built from garbage. A fail verdict with no findings is malformed
// too — the repair round in issue 03 would have nothing to act on.
func TestRunMalformedVerdictFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		verdict string
	}{
		{"not json", `{"status": "fail",`},
		{"unknown status", `{"status": "maybe", "findings": []}`},
		{"fail without findings", `{"status": "fail", "findings": []}`},
		{"finding with empty fields", `{"status": "fail", "findings": [{"location": "", "reasoning": ""}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gh := newFakeGitHub()

			res := runGate(t, gh, writeVerdict(t, tc.verdict))

			if res.Green {
				t.Fatalf("Run returned green for malformed verdict %q", tc.verdict)
			}
			if len(gh.created) != 1 {
				t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
			}
			if !strings.Contains(gh.created[0], "no usable verdict") {
				t.Errorf("comment is not the fail-closed body: %q", gh.created[0])
			}
		})
	}
}

// A re-run on red updates the existing audit comment in place. The
// comment is matched by body prefix including the audit name — never
// by author, and never a comment belonging to a different audit.
func TestRunRedRunUpdatesExistingCommentInPlace(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [{"location": "cmd/partio/status.go:7", "reasoning": "second run reasoning"}]
	}`)
	gh := newFakeGitHub(
		map[string]any{"id": int64(7), "body": "just a human discussing the PR"},
		map[string]any{"id": int64(8), "body": "Minion audit — e2e-need audit failed.\n\nother audit's comment"},
		map[string]any{"id": int64(9), "body": "Minion audit — dead-code audit failed.\n\nfirst run body"},
	)

	res := runGate(t, gh, verdict)

	if res.Green {
		t.Fatalf("Run returned green for a fail verdict")
	}
	if len(gh.created) != 0 {
		t.Fatalf("re-run created a second comment instead of updating: %q", gh.created)
	}
	if len(gh.updated) != 1 {
		t.Fatalf("updated %d comments, want exactly 1: %v", len(gh.updated), gh.updated)
	}
	body, ok := gh.updated[9]
	if !ok {
		t.Fatalf("updated the wrong comment: %v", gh.updated)
	}
	if !strings.Contains(body, "second run reasoning") {
		t.Errorf("updated body does not carry the new findings: %q", body)
	}
}

// Tracer bullet: a fail verdict turns the gate red and posts exactly
// one PR comment carrying the audit prefix, the finding location, and
// the finding reasoning.
func TestRunFailVerdictGoesRedAndPostsOneComment(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [
			{
				"location": "internal/hooks/precommit.go:42",
				"reasoning": "first branch short-circuits on a parameter every remaining call site hardcodes"
			}
		]
	}`)
	gh := newFakeGitHub()

	res := runGate(t, gh, verdict)

	if res.Green {
		t.Fatalf("Run returned green for a fail verdict")
	}
	if len(gh.created) != 1 {
		t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	body := gh.created[0]
	if !strings.HasPrefix(body, "Minion audit — dead-code") {
		t.Errorf("comment does not start with audit prefix: %q", body)
	}
	for _, want := range []string{
		"internal/hooks/precommit.go:42",
		"first branch short-circuits on a parameter every remaining call site hardcodes",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment missing %q:\n%s", want, body)
		}
	}
	if len(gh.updated) != 0 {
		t.Errorf("updated %d comments, want 0", len(gh.updated))
	}
}
