package checksverdict

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/auditgate"
)

// fakeGitHub is the smallest issues-comments double this package
// needs. auditgate has a fuller one for its own tests; here the point
// is only that a checks verdict travels the real gate, so this records
// what the gate posts and nothing else.
type fakeGitHub struct {
	comments []map[string]any
	created  []string
	updated  map[int64]string
	srv      *httptest.Server
	// other records any request outside the comment surface, as
	// "METHOD /path". A push, a merge, or a dispatch would land here.
	other []string
}

func newFakeGitHub() *fakeGitHub {
	return &fakeGitHub{updated: map[int64]string{}}
}

func (f *fakeGitHub) server(t *testing.T) *httptest.Server {
	t.Helper()
	if f.srv != nil {
		return f.srv
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/issues/41/comments", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(f.comments); err != nil {
			t.Errorf("encode comments: %v", err)
		}
	})
	mux.HandleFunc("POST /repos/partio-io/cli/issues/41/comments", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		f.created = append(f.created, body)
		id := int64(100 + len(f.created))
		f.comments = append(f.comments, map[string]any{"id": id, "body": body})
		w.WriteHeader(http.StatusCreated)
		if _, err := fmt.Fprintf(w, `{"id":%d}`, id); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	mux.HandleFunc("PATCH /repos/partio-io/cli/issues/comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		var id int64
		if _, err := fmt.Sscanf(r.PathValue("id"), "%d", &id); err != nil {
			t.Errorf("parse comment id: %v", err)
		}
		f.updated[id] = body
		for _, c := range f.comments {
			if c["id"] == id {
				c["body"] = body
			}
		}
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f.other = append(f.other, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	f.srv = srv
	return srv
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read request body: %v", err)
		return ""
	}
	var payload struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Errorf("decode comment payload: %v", err)
	}
	return payload.Body
}

// fixture returns a captured command output. The files under testdata
// are real `go test` and `golangci-lint` output, not hand-written
// approximations of it.
func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

// runGate sends a verdict through the real gate against the double.
func runGate(t *testing.T, gh *fakeGitHub, verdictPath string) auditgate.Result {
	t.Helper()
	srv := gh.server(t)
	res, err := auditgate.Run(auditgate.Config{
		VerdictPath: verdictPath,
		Repo:        "partio-io/cli",
		PRNumber:    41,
		AuditName:   "checks",
		APIBaseURL:  srv.URL,
		Token:       "test-token",
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatalf("auditgate.Run: %v", err)
	}
	return res
}

// writeTo converts outcomes and writes the verdict, returning its path.
func writeTo(t *testing.T, outcomes ...Outcome) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "verdict.json")
	if err := Write(path, Convert(outcomes)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return path
}

// Re-running the check on a new push replaces its comment instead of
// adding one, and leaves the audits' comments alone. A check that runs
// on every push would otherwise bury the pull request.
func TestRepeatedRunsUpsertOneComment(t *testing.T) {
	path := writeTo(t,
		Outcome{Name: "test", Command: "make test", Output: fixture(t, "test-fail.txt"), Failed: true},
	)
	gh := newFakeGitHub()
	gh.comments = append(gh.comments, map[string]any{
		"id":   int64(8),
		"body": "Minion audit — dead-code audit failed.\n\nanother check's comment",
	})

	for run := 1; run <= 3; run++ {
		if res := runGate(t, gh, path); res.Green {
			t.Fatalf("run %d returned green for a failing test run", run)
		}
	}

	if len(gh.created) != 1 {
		t.Fatalf("three runs created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	if len(gh.updated) != 1 {
		t.Fatalf("three runs updated %d distinct comments, want exactly 1: %v", len(gh.updated), gh.updated)
	}
	if _, ok := gh.updated[101]; !ok {
		t.Errorf("later runs did not update the comment the first run created: %v", gh.updated)
	}
	for _, c := range gh.comments {
		if c["id"] == int64(8) && !strings.Contains(c["body"].(string), "another check's comment") {
			t.Errorf("the checks gate overwrote the dead-code audit's comment: %v", c["body"])
		}
	}
}

// The check has to fail closed exactly as the audits do. A crashed
// converter writes no verdict, and a truncated one writes half of it;
// neither may pass. This is the case that decides whether a green
// check means anything.
func TestMissingOrMalformedVerdictFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		content string // empty means: never write the file at all
	}{
		{"never written", ""},
		{"truncated json", `{"status": "fail", "findings": [`},
		{"unknown status", `{"status": "unknown", "findings": []}`},
		{"fail with no findings", `{"status": "fail", "findings": []}`},
		{"finding with no reasoning", `{"status": "fail", "findings": [{"location": "internal/x.go:1", "reasoning": ""}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "verdict.json")
			if tc.content != "" {
				if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
					t.Fatalf("write verdict: %v", err)
				}
			}
			gh := newFakeGitHub()

			res := runGate(t, gh, path)

			if res.Green {
				t.Fatalf("gate returned green with no usable verdict: %+v", res)
			}
			if len(gh.created) != 1 {
				t.Fatalf("created %d comments, want exactly 1 saying why: %q", len(gh.created), gh.created)
			}
			if !strings.Contains(gh.created[0], "no usable verdict") {
				t.Errorf("comment is not the fail-closed body: %q", gh.created[0])
			}
		})
	}
}

// Everything Convert produces has to survive the round trip to disk
// and back through the gate's own loader. A shape the gate rejects is
// a check that reports nothing while looking like it ran.
func TestWrittenVerdictsAreAlwaysLoadableByTheGate(t *testing.T) {
	cases := []struct {
		name     string
		outcomes []Outcome
		green    bool
	}{
		{"clean", []Outcome{
			{Name: "lint", Command: "make lint", Output: fixture(t, "lint-clean.txt")},
			{Name: "test", Command: "make test", Output: fixture(t, "test-clean.txt")},
		}, true},
		{"lint failed", []Outcome{
			{Name: "lint", Command: "make lint", Output: fixture(t, "lint-fail.txt"), Failed: true},
			{Name: "test", Command: "make test", Output: fixture(t, "test-clean.txt")},
		}, false},
		{"both failed", []Outcome{
			{Name: "lint", Command: "make lint", Output: fixture(t, "lint-fail.txt"), Failed: true},
			{Name: "test", Command: "make test", Output: fixture(t, "test-fail.txt"), Failed: true},
		}, false},
		{"failed with output nobody can parse", []Outcome{
			{Name: "test", Command: "make test", Output: fixture(t, "unparseable.txt"), Failed: true},
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gh := newFakeGitHub()

			res := runGate(t, gh, writeTo(t, tc.outcomes...))

			if res.Green != tc.green {
				t.Fatalf("green = %v, want %v (%s)", res.Green, tc.green, res.Reason)
			}
			if strings.Contains(res.Reason, "fail-closed") {
				t.Errorf("the gate could not read a verdict this package wrote: %s", res.Reason)
			}
		})
	}
}

// Tracer bullet: output from a failing test package travels the whole
// path — parse, verdict file, the audits' own gate — and arrives as a
// red check whose single comment names the package and says what
// failed. Nothing along the way knows the failure came from `make
// test` rather than from an audit session; that is the point of
// reusing the verdict shape.
func TestFailingTestReachesTheGateAsARedCheck(t *testing.T) {
	path := writeTo(t,
		Outcome{Name: "lint", Command: "make lint", Output: fixture(t, "lint-clean.txt")},
		Outcome{Name: "test", Command: "make test", Output: fixture(t, "test-fail.txt"), Failed: true},
	)
	gh := newFakeGitHub()

	res := runGate(t, gh, path)

	if res.Green {
		t.Fatalf("gate returned green for a failing test run: %+v", res)
	}
	if len(gh.created) != 1 {
		t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	body := gh.created[0]
	for _, want := range []string{
		"github.com/partio-io/cli/internal/attribution",
		"TestCalculateSplitsAgentAndHumanLines",
		"Calculate(diff) agent lines = 4, want 5",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment missing %q:\n%s", want, body)
		}
	}
	if len(gh.other) != 0 {
		t.Errorf("the check made requests beyond the comment surface: %q", gh.other)
	}
}
