package review

import (
	"bytes"
	"strings"
	"testing"
)

// A build reviews its issue first. The gate acts as a real run does,
// writes its row marked as a build, and hands the build two outputs:
// blocked after a close, changed after a rewrite. No verdict writes no
// output, so the job fails closed.
func TestBuildGateOutputsPerOutcome(t *testing.T) {
	tests := []struct {
		name    string
		verdict string
		valid   bool
		outputs string
		row     string
		state   string
	}{
		{"keep", keepVerdict, true, "blocked=false\nchanged=false\n", "**keep** (build) ·", "open"},
		{"rewrite", rewriteVerdict, true, "blocked=false\nchanged=true\n", "**rewrite** (build) ·", "open"},
		{"close", closeVerdictFor(ReasonBuilt), true, "blocked=true\n", "**close** (build) ·", "closed"},
		{"no verdict", "not json", false, "", "**no verdict** (build) ·", "open"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			gh.labels = append(gh.labels, ReviewedLabel)
			tracking := gh.trackingIssue(77)
			srv := gh.server(t)

			res, err := Run(Config{
				VerdictPath: writeVerdict(t, tt.verdict),
				Repo:        repo,
				Issue:       12,
				Night:       "2026-10-06",
				Build:       true,
				APIBaseURL:  srv.URL,
				Token:       "test-token",
				HTTPClient:  srv.Client(),
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if res.Valid != tt.valid {
				t.Fatalf("Result = %+v, want Valid %v", res, tt.valid)
			}
			var out bytes.Buffer
			if err := WriteBuildOutputs(&out, res); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.outputs {
				t.Errorf("outputs = %q, want %q", out.String(), tt.outputs)
			}
			if row := gh.comments[tracking][0]["body"].(string); !strings.Contains(row, tt.row) {
				t.Errorf("tracking row is not marked as a build %q:\n%s", tt.row, row)
			}
			if gh.issues[12]["state"] != tt.state {
				t.Errorf("state = %v, want %s", gh.issues[12]["state"], tt.state)
			}
			if acted := len(gh.comments[12]) > 0; acted != tt.valid {
				t.Errorf("evidence comment posted = %v, want %v", acted, tt.valid)
			}
		})
	}
}
