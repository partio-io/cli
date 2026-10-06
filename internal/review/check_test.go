package review

import (
	"path/filepath"
	"strings"
	"testing"
)

// Check runs the checks of the gate on a verdict and changes nothing: it
// reads the issue, and it sends no other request. The review program runs
// it on its own verdict, so a cause it misses is a "no verdict" row later.
func TestCheckReportsWhatTheGateWouldRefuse(t *testing.T) {
	noClaims := `"premise": {"verdict": "no-claims", "claims": []}`
	if !strings.Contains(rewriteVerdict, noClaims) {
		t.Fatalf("rewriteVerdict no longer carries %s", noClaims)
	}
	failing := strings.Replace(rewriteVerdict, noClaims,
		`"premise": {"verdict": "fails", "claims": [{"claim": "the hook runs once", "evidence": "internal/hooks", "verdict": "fails", "excerpt": "run()"}]}`, 1)
	corrected := strings.Replace(failing, `"excerpt": "run()"}`, `"excerpt": "run()", "correction": "the hook runs twice"}`, 1)
	noBlock := strings.Replace(rewriteVerdict, "<!-- partio:premise:v1 -->", "", 1)

	tests := []struct {
		name    string
		verdict string // empty: no verdict file
		cause   string // empty: the verdict passes
	}{
		{"valid keep", keepVerdict, ""},
		{"valid rewrite", rewriteVerdict, ""},
		{"no verdict file", "", "read verdict"},
		{"failing claim without a correction", failing, "claim 1 fails and has no correction"},
		{"rewrite with a corrected failing claim", corrected, "rewrite with a failing premise"},
		{"rewrite whose body has no premise block", noBlock, "rewrite body has no premise block"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			srv := gh.server(t)
			path := filepath.Join(t.TempDir(), "verdict.json")
			if tt.verdict != "" {
				path = writeVerdict(t, tt.verdict)
			}

			cause, err := Check(Config{
				VerdictPath: path,
				Repo:        repo,
				Issue:       12,
				APIBaseURL:  srv.URL,
				Token:       "test-token",
				HTTPClient:  srv.Client(),
			})

			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if tt.cause == "" && cause != "" {
				t.Errorf("Check refused a valid verdict: %s", cause)
			}
			if tt.cause != "" && !strings.Contains(cause, tt.cause) {
				t.Errorf("cause = %q, want it to contain %q", cause, tt.cause)
			}
			for _, r := range gh.requests {
				if !strings.HasPrefix(r, "GET ") {
					t.Errorf("Check changed something: %s", r)
				}
			}
		})
	}
}
