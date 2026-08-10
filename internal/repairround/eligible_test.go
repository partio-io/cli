package repairround

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEligible(t *testing.T) {
	const base = "partio-io/cli"

	tests := []struct {
		name        string
		labels      []string
		headRepo    string
		wantAllowed bool
		wantReason  string // a fragment the reason must name
	}{
		{
			name:        "labelled branch in this repository",
			labels:      []string{"minion"},
			headRepo:    base,
			wantAllowed: true,
			wantReason:  "minion",
		},
		{
			name:        "labelled among other labels",
			labels:      []string{"enhancement", "minion", "needs-review"},
			headRepo:    base,
			wantAllowed: true,
			wantReason:  "minion",
		},
		{
			name:        "no labels at all",
			labels:      nil,
			headRepo:    base,
			wantAllowed: false,
			wantReason:  "minion",
		},
		{
			name:        "labelled, but not with the repair label",
			labels:      []string{"bug", "documentation"},
			headRepo:    base,
			wantAllowed: false,
			wantReason:  "minion",
		},
		// The operator's own branch is the case this rule exists for:
		// a hand-written pull request gets the red or green signal and
		// no bot commit on top of work a person wrote.
		{
			name:        "a near miss is not the label",
			labels:      []string{"minions", "Minion"},
			headRepo:    base,
			wantAllowed: false,
			wantReason:  "minion",
		},
		// A fork head cannot be pushed to, so the refusal has to come
		// before anything is fetched — not as a push that fails.
		{
			name:        "head branch lives in a fork",
			labels:      []string{"minion"},
			headRepo:    "someone-else/cli",
			wantAllowed: false,
			wantReason:  "someone-else/cli",
		},
		{
			name:        "fork head without the label",
			labels:      nil,
			headRepo:    "someone-else/cli",
			wantAllowed: false,
			wantReason:  "minion",
		},
		// GitHub reports no head repository for a deleted fork. An
		// unknown head is refused rather than assumed to be ours.
		{
			name:        "head repository unknown",
			labels:      []string{"minion"},
			headRepo:    "",
			wantAllowed: false,
			wantReason:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Eligible(tt.labels, tt.headRepo, base)

			if got.Allowed != tt.wantAllowed {
				t.Errorf("Eligible(%q, %q, %q).Allowed = %t, want %t",
					tt.labels, tt.headRepo, base, got.Allowed, tt.wantAllowed)
			}
			if got.Reason == "" {
				t.Fatal("Eligible returned no reason; the workflow log is the only place this decision is visible")
			}
			if !strings.Contains(got.Reason, tt.wantReason) {
				t.Errorf("reason %q does not name %q", got.Reason, tt.wantReason)
			}
		})
	}
}

// The base repository is the one the workflow runs in, so a head that
// matches it is pushable whatever either is called.
func TestEligibleComparesHeadAgainstTheGivenBase(t *testing.T) {
	got := Eligible([]string{"minion"}, "other-org/other-name", "other-org/other-name")

	if !got.Allowed {
		t.Errorf("Eligible refused a head in the base repository: %q", got.Reason)
	}
}

// fakePR serves one pull request. The labels and the head repository
// come from the API rather than from a workflow expression, so the
// decision has one implementation and the workflow carries no label
// logic of its own.
func fakePR(t *testing.T, labels []string, headRepo string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/pulls/41", func(w http.ResponseWriter, _ *http.Request) {
		body := map[string]any{"labels": []map[string]any{}, "head": map[string]any{}}
		named := []map[string]any{}
		for _, l := range labels {
			named = append(named, map[string]any{"name": l})
		}
		body["labels"] = named
		if headRepo != "" {
			body["head"] = map[string]any{"repo": map[string]any{"full_name": headRepo}}
		}
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("encode pull request: %v", err)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestRunEligibility(t *testing.T) {
	tests := []struct {
		name        string
		labels      []string
		headRepo    string
		wantAllowed bool
	}{
		{name: "minion pull request in this repository", labels: []string{"minion"}, headRepo: "partio-io/cli", wantAllowed: true},
		{name: "hand-written pull request", labels: []string{"bug"}, headRepo: "partio-io/cli"},
		{name: "fork head", labels: []string{"minion"}, headRepo: "someone-else/cli"},
		{name: "no head repository", labels: []string{"minion"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := fakePR(t, tt.labels, tt.headRepo)

			got, err := RunEligibility(Config{
				Repo:       "partio-io/cli",
				PRNumber:   41,
				Check:      "checks",
				APIBaseURL: srv.URL,
				Token:      "test-token",
				HTTPClient: srv.Client(),
			})
			if err != nil {
				t.Fatalf("RunEligibility: %v", err)
			}

			if got.Allowed != tt.wantAllowed {
				t.Errorf("RunEligibility = %+v, want allowed %t", got, tt.wantAllowed)
			}
			if got.Reason == "" {
				t.Error("RunEligibility returned no reason")
			}
		})
	}
}

// A pull request the API will not describe leaves the decision unknown.
// Guessing "repairable" there would push a commit on no evidence, so
// the error travels up and the caller declines.
func TestRunEligibilityReportsAnUnreadablePullRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/pulls/41", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := RunEligibility(Config{
		Repo:       "partio-io/cli",
		PRNumber:   41,
		Check:      "checks",
		APIBaseURL: srv.URL,
		Token:      "test-token",
		HTTPClient: srv.Client(),
	})
	if err == nil {
		t.Fatal("RunEligibility accepted a pull request it could not read")
	}
}
