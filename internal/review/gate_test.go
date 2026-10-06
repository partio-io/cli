package review

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const repo = "partio-io/cli"

// fakeGitHub is an issues, labels and comments API double for one
// repository. It records every request as "METHOD /path" in order, so
// a test can assert the exact requests a gate run made.
type fakeGitHub struct {
	issues   map[int]map[string]any
	labels   []string
	comments map[int][]map[string]any
	nextID   int64
	requests []string
	srv      *httptest.Server
}

func newFakeGitHub() *fakeGitHub {
	return &fakeGitHub{
		issues: map[int]map[string]any{
			12: issueJSON(12, "Add a retry to the pre-push hook", "body", nil),
			13: issueJSON(13, "Cache the session index", "body", nil),
		},
		comments: map[int][]map[string]any{},
		nextID:   500,
	}
}

func issueJSON(number int, title, body string, labels []string) map[string]any {
	ls := []map[string]any{}
	for _, l := range labels {
		ls = append(ls, map[string]any{"name": l})
	}
	return map[string]any{
		"number":   number,
		"title":    title,
		"body":     body,
		"state":    "open",
		"labels":   ls,
		"html_url": fmt.Sprintf("https://github.com/%s/issues/%d", repo, number),
	}
}

func (f *fakeGitHub) server(t *testing.T) *httptest.Server {
	t.Helper()
	if f.srv != nil {
		return f.srv
	}
	base := "/repos/" + repo
	write := func(w http.ResponseWriter, status int, v any) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}
	decode := func(r *http.Request, v any) {
		if err := json.NewDecoder(r.Body).Decode(v); err != nil {
			t.Errorf("decode %s %s: %v", r.Method, r.URL.Path, err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+base+"/issues/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.PathValue("n"))
		issue, ok := f.issues[n]
		if !ok {
			write(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
			return
		}
		write(w, http.StatusOK, issue)
	})
	mux.HandleFunc("GET "+base+"/issues", func(w http.ResponseWriter, r *http.Request) {
		label := r.URL.Query().Get("labels")
		if r.URL.Query().Get("state") != "open" {
			t.Errorf("issue list without state=open: %s", r.URL.RawQuery)
		}
		var out []map[string]any
		for _, n := range slices.Sorted(maps.Keys(f.issues)) {
			issue := f.issues[n]
			if issue["state"] != "open" {
				continue
			}
			for _, l := range issue["labels"].([]map[string]any) {
				if l["name"] == label {
					out = append(out, issue)
				}
			}
		}
		if out == nil {
			out = []map[string]any{}
		}
		write(w, http.StatusOK, out)
	})
	mux.HandleFunc("POST "+base+"/issues", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Title  string   `json:"title"`
			Body   string   `json:"body"`
			Labels []string `json:"labels"`
		}
		decode(r, &p)
		n := 900 + len(f.issues)
		f.issues[n] = issueJSON(n, p.Title, p.Body, p.Labels)
		write(w, http.StatusCreated, f.issues[n])
	})
	mux.HandleFunc("GET "+base+"/labels/{name}", func(w http.ResponseWriter, r *http.Request) {
		if !slices.Contains(f.labels, r.PathValue("name")) {
			write(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
			return
		}
		write(w, http.StatusOK, map[string]string{"name": r.PathValue("name")})
	})
	mux.HandleFunc("POST "+base+"/labels", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			Name string `json:"name"`
		}
		decode(r, &p)
		f.labels = append(f.labels, p.Name)
		write(w, http.StatusCreated, p)
	})
	mux.HandleFunc("GET "+base+"/issues/{n}/comments", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.PathValue("n"))
		cs := f.comments[n]
		if cs == nil {
			cs = []map[string]any{}
		}
		write(w, http.StatusOK, cs)
	})
	mux.HandleFunc("POST "+base+"/issues/{n}/comments", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.PathValue("n"))
		var p struct {
			Body string `json:"body"`
		}
		decode(r, &p)
		f.nextID++
		c := map[string]any{"id": f.nextID, "body": p.Body}
		f.comments[n] = append(f.comments[n], c)
		write(w, http.StatusCreated, c)
	})
	mux.HandleFunc("PATCH "+base+"/issues/comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var p struct {
			Body string `json:"body"`
		}
		decode(r, &p)
		for _, cs := range f.comments {
			for _, c := range cs {
				if c["id"] == id {
					c["body"] = p.Body
					write(w, http.StatusOK, c)
					return
				}
			}
		}
		write(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests = append(f.requests, r.Method+" "+r.URL.Path)
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	f.srv = srv
	return srv
}

// trackingIssue seeds an existing tracking issue and returns its number.
func (f *fakeGitHub) trackingIssue(number int) int {
	f.labels = append(f.labels, TrackingLabel)
	f.issues[number] = issueJSON(number, "Minion review log", TrackingMarker+"\nlog", []string{TrackingLabel})
	return number
}

func writeVerdict(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write verdict: %v", err)
	}
	return path
}

func runGate(t *testing.T, gh *fakeGitHub, issue int, verdictPath, night string) Result {
	t.Helper()
	srv := gh.server(t)
	res, err := Run(Config{
		VerdictPath: verdictPath,
		Repo:        repo,
		Issue:       issue,
		Night:       night,
		APIBaseURL:  srv.URL,
		Token:       "test-token",
		HTTPClient:  srv.Client(),
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res
}

const keepVerdict = `{
	"issue": 12,
	"outcome": "keep",
	"premise": {
		"verdict": "holds",
		"claims": [
			{
				"claim": "the pre-push hook has no retry",
				"evidence": "internal/hooks/prepush.go",
				"verdict": "holds",
				"excerpt": "return push(remote)"
			}
		]
	},
	"fit": {"applies": true, "reason": "pre-push is in scope"},
	"built": {"built": false, "evidence": ""},
	"duplicates": []
}`

// A valid keep verdict on a first run creates the label and the tracking
// issue, writes one night comment with one row, and touches the
// reviewed issue only to read its title.
func TestGateDryRunValidVerdictWritesOneRow(t *testing.T) {
	gh := newFakeGitHub()

	res := runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-06")

	if !res.Valid {
		t.Fatalf("Run returned no verdict for a valid keep: %s", res.Cause)
	}
	want := []string{
		"GET /repos/partio-io/cli/issues/12",
		"GET /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/labels/minion-review-log",
		"POST /repos/partio-io/cli/labels",
		"POST /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/issues/902/comments",
		"POST /repos/partio-io/cli/issues/902/comments",
	}
	if !slices.Equal(gh.requests, want) {
		t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
	}
	created := gh.issues[902]
	if !strings.HasPrefix(created["body"].(string), TrackingMarker+"\n") {
		t.Errorf("tracking issue body does not start with the marker: %q", created["body"])
	}
	comments := gh.comments[902]
	if len(comments) != 1 {
		t.Fatalf("tracking issue has %d comments, want 1", len(comments))
	}
	body := comments[0]["body"].(string)
	if !strings.HasPrefix(body, NightMarker("2026-10-06")) {
		t.Errorf("night comment does not start with its marker: %q", body)
	}
	for _, want := range []string{"#12", "Add a retry to the pre-push hook", "keep", "https://github.com/partio-io/cli/issues/12"} {
		if !strings.Contains(body, want) {
			t.Errorf("night comment missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "\n- "); got != 1 {
		t.Errorf("night comment has %d rows, want 1:\n%s", got, body)
	}
}
