package repairround

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeGitHub is a minimal pull-request-commits API double. It serves
// the commit messages it is seeded with on the first page and an empty
// page after that, so the pagination loop terminates.
type fakeGitHub struct {
	messages []string
	pages    int
}

func (f *fakeGitHub) server(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/pulls/41/commits", func(w http.ResponseWriter, r *http.Request) {
		f.pages++
		body := []map[string]any{}
		if page := r.URL.Query().Get("page"); page == "" || page == "1" {
			for _, m := range f.messages {
				body = append(body, map[string]any{"commit": map[string]any{"message": m}})
			}
		}
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("encode commits: %v", err)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// The tracer bullet: a branch carrying one dead-code repair, one
// repair for a different check, and human work reads as "round 2 of
// dead-code may start".
func TestRunCountsRepairsForOneCheckOnly(t *testing.T) {
	fake := &fakeGitHub{messages: []string{
		"feat: add the thing",
		Subject("dead-code", 41),
		Subject("lint-test", 41),
		"fix: address review\n\nA body that mentions audit(dead-code): markers.",
	}}
	srv := fake.server(t)

	got, err := Run(Config{
		Repo:       "partio-io/cli",
		PRNumber:   41,
		Check:      "dead-code",
		MaxRounds:  DefaultMaxRounds,
		APIBaseURL: srv.URL,
		Token:      "test-token",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := Decision{Allowed: true, Round: 2, Prior: 1}
	if got != want {
		t.Errorf("Run = %+v, want %+v", got, want)
	}
}
