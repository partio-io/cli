package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDupesPrintsTheCandidatesAsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/partio-io/cli/issues" || r.URL.Query().Get("state") != "all" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"number": 681, "title": "Add Goose coding agent integration", "state": "open",
			 "body": "Origin: entireio/cli#2075", "labels": [{"name": "minion-proposal"}]},
			{"number": 688, "title": "Add Goose coding agent integration", "state": "closed", "state_reason": "completed",
			 "body": "Inspired by entireio/cli PR #2075, #2077", "labels": [{"name": "minion-proposal"}]}
		]`))
	}))
	defer srv.Close()
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", srv.URL)

	var out bytes.Buffer
	code := dupes([]string{
		"--source", "entireio/cli#2075", "--source", "https://github.com/entireio/cli/pull/2077",
		"--title", "Add Goose coding agent integration", "--exclude", "681",
	}, &out)
	if code != 0 {
		t.Fatalf("dupes exit %d, output %q", code, out.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not a JSON list: %v\n%s", err, out.String())
	}
	if len(got) != 1 || got[0]["number"] != float64(688) || got[0]["match"] != "source" ||
		got[0]["state_reason"] != "completed" {
		t.Errorf("dupes printed %s", out.String())
	}
}

func TestDupesWithoutAMatchPrintsAnEmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", srv.URL)

	var out bytes.Buffer
	if code := dupes([]string{"--title", "Anything"}, &out); code != 0 {
		t.Fatalf("dupes exit %d", code)
	}
	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("dupes printed %q, want []", out.String())
	}
}

func TestDupesNeedsATitle(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	var out bytes.Buffer
	if code := dupes([]string{"--source", "entireio/cli#1"}, &out); code != 2 {
		t.Errorf("dupes without --title exit %d, want 2", code)
	}
}

// fakeGateGitHub serves what one gate run reads, with a tracking issue
// in place, and records "METHOD /path" for every request.
func fakeGateGitHub(t *testing.T) *[]string {
	t.Helper()
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/partio-io/cli/issues/12":
			_, _ = w.Write([]byte(`{"number": 12, "title": "Retry the push", "state": "open", "html_url": "https://github.com/partio-io/cli/issues/12",
				"labels": [{"name": "minion-proposal"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/partio-io/cli/issues":
			_, _ = w.Write([]byte(`[{"number": 77, "title": "Minion review log", "body": "<!-- minion-review-log -->\nlog"}]`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/comments"), r.URL.Path == "/repos/partio-io/cli/pulls":
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", srv.URL)
	return &got
}

// The sweep passes --dry-run="$DRY_RUN": false acts on the reviewed
// issue, true only writes the tracking row.
func TestGatePassesTheDryRunValue(t *testing.T) {
	verdict := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(verdict, []byte(`{"issue": 12, "outcome": "keep",
		"premise": {"verdict": "no-claims", "claims": []},
		"fit": {"applies": true, "reason": "in scope"},
		"built": {"built": false, "evidence": ""}, "duplicates": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		dryRun string
		acts   bool
	}{{"false", true}, {"true", false}} {
		t.Run("dry-run="+tt.dryRun, func(t *testing.T) {
			got := fakeGateGitHub(t)

			code := gate([]string{"--issue", "12", "--verdict", verdict, "--dry-run=" + tt.dryRun, "--night", "2026-10-06"})

			if code != 0 {
				t.Fatalf("gate exit %d, requests %q", code, *got)
			}
			if acted := slices.Contains(*got, "POST /repos/partio-io/cli/issues/12/labels"); acted != tt.acts {
				t.Errorf("acted on #12 = %v, want %v; requests %q", acted, tt.acts, *got)
			}
		})
	}
}

func TestNextPrintsOneIssuePerLine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/partio-io/cli/issues" || r.URL.Query().Get("labels") != "minion-proposal" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`[
			{"number": 40, "title": "b", "state": "open", "created_at": "2026-09-02T10:00:00Z",
			 "labels": [{"name": "minion-proposal"}]},
			{"number": 31, "title": "a", "state": "open", "created_at": "2026-08-01T10:00:00Z",
			 "labels": [{"name": "minion-proposal"}]}
		]`))
	}))
	defer srv.Close()
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", srv.URL)

	var out bytes.Buffer
	if code := next(nil, &out); code != 0 {
		t.Fatalf("next exit %d, output %q", code, out.String())
	}
	if out.String() != "31\n40\n" {
		t.Errorf("next printed %q, want %q", out.String(), "31\n40\n")
	}
}

func TestNextRefusesABadSample(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "partio-io/cli")
	t.Setenv("GH_TOKEN", "tok")
	for _, args := range [][]string{{"--sample", "0"}, {"--sample=-3"}, {"--sample", "x"}} {
		var out bytes.Buffer
		if code := next(args, &out); code != 2 {
			t.Errorf("next %q exit %d, want 2", args, code)
		}
	}
}
