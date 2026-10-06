package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
