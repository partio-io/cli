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

// buildVerdicts are the verdict files the build-mode tests run the gate
// on: one per outcome, and one that is no verdict.
var buildVerdicts = map[string]string{
	"keep": `{"issue": 12, "outcome": "keep",
		"premise": {"verdict": "no-claims", "claims": []},
		"fit": {"applies": true, "reason": "in scope"},
		"built": {"built": false, "evidence": ""}, "duplicates": []}`,
	"rewrite": `{"issue": 12, "outcome": "rewrite",
		"premise": {"verdict": "no-claims", "claims": []},
		"fit": {"applies": true, "reason": "in scope"},
		"built": {"built": false, "evidence": ""}, "duplicates": [],
		"rewrite": {"title": "Retry the push once", "changes": ["added the premise block"],
			"body": "## What\n\nRetry the push once.\n\n## Premise\n\n<!-- partio:premise:v1 -->\n\n- the pre-push hook has no retry [evidence: ` + "`internal/hooks/prepush.go`" + `]\n\n## Acceptance Criteria\n\n- [ ] a push that fails runs once more\n\nProposal id: retry-pre-push\n"}}`,
	"close": `{"issue": 12, "outcome": "close", "close_reason": "built",
		"premise": {"verdict": "no-claims", "claims": []},
		"fit": {"applies": true, "reason": "in scope"},
		"built": {"built": true, "evidence": "internal/hooks/retry.go already retries"}, "duplicates": []}`,
	"no verdict": `not json`,
}

// A build reviews its issue with --mode build: the gate acts on the
// issue as a real run, and appends the blocked and changed step outputs
// the build workflow reads to GITHUB_OUTPUT. No verdict exits 1 and
// writes no output, so the job fails closed.
func TestGateBuildModeWritesTheStepOutputs(t *testing.T) {
	for _, tt := range []struct {
		outcome string
		code    int
		outputs string
		acted   bool
	}{
		{"keep", 0, "blocked=false\nchanged=false\n", true},
		{"rewrite", 0, "blocked=false\nchanged=true\n", true},
		{"close", 0, "blocked=true\n", true},
		{"no verdict", 1, "", false},
	} {
		t.Run(tt.outcome, func(t *testing.T) {
			verdict := filepath.Join(t.TempDir(), "verdict.json")
			if err := os.WriteFile(verdict, []byte(buildVerdicts[tt.outcome]), 0o644); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(t.TempDir(), "output")
			if err := os.WriteFile(output, []byte("earlier=1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GITHUB_OUTPUT", output)
			got := fakeGateGitHub(t)

			code := gate([]string{"--mode", "build", "--issue", "12", "--verdict", verdict})

			if code != tt.code {
				t.Fatalf("gate exit %d, want %d; requests %q", code, tt.code, *got)
			}
			if acted := slices.Contains(*got, "POST /repos/partio-io/cli/issues/12/labels"); acted != tt.acted {
				t.Errorf("acted on #12 = %v, want %v; requests %q", acted, tt.acted, *got)
			}
			out, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if want := "earlier=1\n" + tt.outputs; string(out) != want {
				t.Errorf("GITHUB_OUTPUT = %q, want %q", out, want)
			}
		})
	}
}

// The gate can fail on the tracking issue after it acted. A build still
// gets its outputs then, so the failure step does not mark an issue the
// gate closed as a failed build.
func TestGateBuildModeWritesOutputsWhenTrackingFailsAfterActing(t *testing.T) {
	verdict := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(verdict, []byte(`{"issue": 12, "outcome": "keep",
		"premise": {"verdict": "no-claims", "claims": []},
		"fit": {"applies": true, "reason": "in scope"},
		"built": {"built": false, "evidence": ""}, "duplicates": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "output")
	t.Setenv("GITHUB_OUTPUT", output)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/partio-io/cli/issues/12":
			_, _ = w.Write([]byte(`{"number": 12, "title": "Retry the push", "state": "open", "html_url": "https://github.com/partio-io/cli/issues/12"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/partio-io/cli/issues":
			w.WriteHeader(http.StatusInternalServerError) // the tracking issue lookup fails
		case r.Method == http.MethodGet && (strings.HasSuffix(r.URL.Path, "/comments") || r.URL.Path == "/repos/partio-io/cli/pulls"):
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

	if code := gate([]string{"--mode", "build", "--issue", "12", "--verdict", verdict}); code != 1 {
		t.Fatalf("gate exit %d, want 1", code)
	}
	out, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if want := "blocked=false\nchanged=false\n"; string(out) != want {
		t.Errorf("GITHUB_OUTPUT = %q, want %q", out, want)
	}
}

// Build mode acts, so it refuses --dry-run, and it needs GITHUB_OUTPUT
// to hand its outputs to the workflow.
func TestGateBuildModeUsageErrors(t *testing.T) {
	verdict := filepath.Join(t.TempDir(), "verdict.json")
	for _, tt := range []struct {
		name   string
		args   []string
		output string
	}{
		{"dry run", []string{"--mode", "build", "--dry-run", "--issue", "12", "--verdict", verdict}, filepath.Join(t.TempDir(), "output")},
		{"no GITHUB_OUTPUT", []string{"--mode", "build", "--issue", "12", "--verdict", verdict}, ""},
		{"unknown mode", []string{"--mode", "nightly", "--issue", "12", "--verdict", verdict}, filepath.Join(t.TempDir(), "output")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_OUTPUT", tt.output)
			got := fakeGateGitHub(t)
			if code := gate(tt.args); code != 2 {
				t.Errorf("gate exit %d, want 2", code)
			}
			if len(*got) != 0 {
				t.Errorf("a usage error reached GitHub: %q", *got)
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
