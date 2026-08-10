package auditgate

import (
	"encoding/json"
	"fmt"
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
	srv      *httptest.Server
	// other records every request that is not a comment read or
	// write, as "METHOD /path". Closing a pull request, reopening it,
	// or dispatching a rebuild would land here.
	other []string
}

func newFakeGitHub(seed ...map[string]any) *fakeGitHub {
	return &fakeGitHub{comments: seed, updated: map[int64]string{}}
}

// server starts the double once and reuses it, so several gate runs
// can share one pull request and see each other's comments.
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
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read POST body: %v", err)
			return
		}
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode created comment: %v", err)
		}
		f.created = append(f.created, payload.Body)
		id := int64(100 + len(f.created))
		f.comments = append(f.comments, map[string]any{"id": id, "body": payload.Body})
		w.WriteHeader(http.StatusCreated)
		if _, err := fmt.Fprintf(w, `{"id":%d}`, id); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	mux.HandleFunc("PATCH /repos/partio-io/cli/issues/comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read PATCH body: %v", err)
			return
		}
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
		for _, c := range f.comments {
			if c["id"] == id {
				c["body"] = payload.Body
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
	return runGateRound(t, gh, verdictPath, 0, 0)
}

// runGateRound runs the gate as a numbered repair round. Round and
// maxRounds of zero mean the caller has no round accounting, which is
// how the audits that predate the round counter still call it.
func runGateRound(t *testing.T, gh *fakeGitHub, verdictPath string, round, maxRounds int) Result {
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
		Round:       round,
		MaxRounds:   maxRounds,
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

// Tracer bullet for the spent budget: the last round's verdict still
// fails, so the gate's one comment has to carry both halves of the
// story — which findings survived, and that three rounds were spent
// reaching them.
func TestRunLastRoundReportsExhaustedBudget(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [
			{
				"location": "internal/session/state.go:88",
				"reasoning": "reachable only from a test helper the PR deleted"
			}
		]
	}`)
	gh := newFakeGitHub()

	res := runGateRound(t, gh, verdict, 3, 3)

	if res.Green {
		t.Fatalf("Run returned green for a fail verdict on the last round")
	}
	if len(gh.created) != 1 {
		t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	body := gh.created[0]
	if !strings.HasPrefix(body, "Minion audit — dead-code") {
		t.Errorf("comment does not start with audit prefix, so the next upsert cannot find it: %q", body)
	}
	for _, want := range []string{
		"internal/session/state.go:88",
		"reachable only from a test helper the PR deleted",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("exhausted comment missing %q:\n%s", want, body)
		}
	}
	if !strings.Contains(body, "3 repair round") {
		t.Errorf("exhausted comment does not say how many rounds ran:\n%s", body)
	}
}

// Every survivor is named, not just the first: the comment is the
// whole handover to whoever picks the pull request up.
func TestRunExhaustedCommentNamesEverySurvivingFinding(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [
			{"location": "internal/git/branch.go:31", "reasoning": "only caller was deleted by this PR"},
			{"location": "internal/git/diff.go:77", "reasoning": "unreachable after the helper merge"},
			{"location": "internal/config/env.go:9", "reasoning": "superseded by the layered loader"}
		]
	}`)
	gh := newFakeGitHub()

	runGateRound(t, gh, verdict, 3, 3)

	if len(gh.created) != 1 {
		t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	body := gh.created[0]
	for _, want := range []string{
		"internal/git/branch.go:31", "only caller was deleted by this PR",
		"internal/git/diff.go:77", "unreachable after the helper merge",
		"internal/config/env.go:9", "superseded by the layered loader",
		"3 findings survived",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("exhausted comment missing %q:\n%s", want, body)
		}
	}
}

// The round count is the operator's cue for how much was already
// tried, so the comment must carry the round this run actually is,
// not the cap and not a constant.
func TestRunExhaustedCommentStatesRoundsRun(t *testing.T) {
	cases := []struct {
		name      string
		round     int
		maxRounds int
		want      string
	}{
		{"cap of three", 3, 3, "3 repair round"},
		{"cap of one", 1, 1, "1 repair round"},
		{"round past the cap", 4, 3, "4 repair round"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict := writeVerdict(t, `{
				"status": "fail",
				"findings": [{"location": "internal/git/diff.go:12", "reasoning": "no caller remains"}]
			}`)
			gh := newFakeGitHub()

			res := runGateRound(t, gh, verdict, tc.round, tc.maxRounds)

			if res.Green {
				t.Fatalf("Run returned green for a fail verdict")
			}
			if len(gh.created) != 1 {
				t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
			}
			if !strings.Contains(gh.created[0], tc.want) {
				t.Errorf("comment does not state %q:\n%s", tc.want, gh.created[0])
			}
		})
	}
}

// "Gave up after spending the budget" and "failed on the first pass"
// are the two red states an operator has to tell apart. A run with no
// round accounting reads as the latter, which is how the audits that
// predate the counter keep their existing comment.
func TestRunFailBodyDistinguishesFirstPassFromExhausted(t *testing.T) {
	const verdictJSON = `{
		"status": "fail",
		"findings": [{"location": "internal/config/load.go:30", "reasoning": "unreferenced since the PR"}]
	}`
	cases := []struct {
		name       string
		round      int
		maxRounds  int
		wantRounds bool
	}{
		{"first round of three", 1, 3, false},
		{"middle round of three", 2, 3, false},
		{"no round accounting", 0, 0, false},
		{"last round of three", 3, 3, true},
	}
	bodies := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gh := newFakeGitHub()

			runGateRound(t, gh, writeVerdict(t, verdictJSON), tc.round, tc.maxRounds)

			if len(gh.created) != 1 {
				t.Fatalf("created %d comments, want exactly 1: %q", len(gh.created), gh.created)
			}
			body := gh.created[0]
			bodies[tc.name] = body
			if got := strings.Contains(body, "repair budget is spent"); got != tc.wantRounds {
				t.Errorf("body reports a spent budget = %v, want %v:\n%s", got, tc.wantRounds, body)
			}
		})
	}
	if bodies["first round of three"] == bodies["last round of three"] {
		t.Errorf("first-pass and exhausted comments are identical:\n%s", bodies["first round of three"])
	}
	if bodies["no round accounting"] != bodies["first round of three"] {
		t.Errorf("a run with no round accounting changed the existing fail comment:\n%s\nvs\n%s",
			bodies["no round accounting"], bodies["first round of three"])
	}
}

// Three rounds on one pull request leave one comment, not three. The
// exhausted body opens differently from the fail body, so this also
// proves the prefix match still finds the earlier round's comment and
// replaces it rather than appending beside it.
func TestRunRepeatedRoundsUpsertOneComment(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [{"location": "internal/attribution/lines.go:64", "reasoning": "dead since the rewrite"}]
	}`)
	gh := newFakeGitHub(map[string]any{"id": int64(8), "body": "Minion audit — e2e-need audit failed.\n\nanother check"})

	for round := 1; round <= 3; round++ {
		if res := runGateRound(t, gh, verdict, round, 3); res.Green {
			t.Fatalf("round %d returned green for a fail verdict", round)
		}
	}

	if len(gh.created) != 1 {
		t.Fatalf("three rounds created %d comments, want exactly 1: %q", len(gh.created), gh.created)
	}
	if len(gh.updated) != 1 {
		t.Fatalf("three rounds updated %d distinct comments, want exactly 1: %v", len(gh.updated), gh.updated)
	}
	final, ok := gh.updated[101]
	if !ok {
		t.Fatalf("later rounds did not update the comment the first round created: %v", gh.updated)
	}
	if !strings.Contains(final, "repair budget is spent") {
		t.Errorf("the surviving comment is not the last round's:\n%s", final)
	}
	for _, c := range gh.comments {
		if c["id"] == int64(8) && c["body"] != "Minion audit — e2e-need audit failed.\n\nanother check" {
			t.Errorf("the repair rounds overwrote another check's comment: %v", c["body"])
		}
	}
}

// A spent budget stops the pipeline and reports. It must not close the
// pull request, reopen it, or dispatch a rebuild: the gate's only
// side effect is the comment, and its only verdict is red.
func TestRunExhaustedLeavesPullRequestRedAndOpen(t *testing.T) {
	verdict := writeVerdict(t, `{
		"status": "fail",
		"findings": [{"location": "internal/log/setup.go:20", "reasoning": "no remaining caller"}]
	}`)
	gh := newFakeGitHub()

	res := runGateRound(t, gh, verdict, 3, 3)

	if res.Green {
		t.Fatalf("Run returned green for an exhausted budget: %+v", res)
	}
	if len(gh.other) != 0 {
		t.Errorf("the gate made requests beyond the comment surface: %q", gh.other)
	}
}
