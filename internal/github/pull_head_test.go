package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestGetPullHead(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantHead *PullRepo
		wantSHA  string
		wantBase string
		wantN    int
	}{
		{
			name:     "decodes head and base",
			body:     `{"commits":312,"head":{"sha":"abc","repo":{"full_name":"f/r","clone_url":"https://github.com/f/r.git"}},"base":{"repo":{"full_name":"o/r"}}}`,
			wantHead: &PullRepo{FullName: "f/r", CloneURL: "https://github.com/f/r.git"},
			wantSHA:  "abc",
			wantBase: "o/r",
			wantN:    312,
		},
		{
			name:     "decodes a null head repository",
			body:     `{"head":{"sha":"abc","repo":null},"base":{"repo":{"full_name":"o/r"}}}`,
			wantSHA:  "abc",
			wantBase: "o/r",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/o/r/pulls/7" {
					http.NotFound(w, r)
					return
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			pr, err := Client{BaseURL: srv.URL, HTTPClient: srv.Client()}.GetPullHead("o/r", 7)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (pr.Head.Repo == nil) != (tt.wantHead == nil) || (pr.Head.Repo != nil && *pr.Head.Repo != *tt.wantHead) {
				t.Errorf("head repo = %+v, want %+v", pr.Head.Repo, tt.wantHead)
			}
			if pr.Head.SHA != tt.wantSHA {
				t.Errorf("head sha = %q, want %q", pr.Head.SHA, tt.wantSHA)
			}
			if pr.Base.Repo.FullName != tt.wantBase {
				t.Errorf("base repo = %q, want %q", pr.Base.Repo.FullName, tt.wantBase)
			}
			if pr.Commits != tt.wantN {
				t.Errorf("commits = %d, want %d", pr.Commits, tt.wantN)
			}
		})
	}
}

func TestListPullCommits(t *testing.T) {
	const total = 250
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls/7/commits" {
			http.NotFound(w, r)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		start := min((page-1)*PerPage, total)
		end := min(start+PerPage, total)
		var commits []map[string]any
		for i := start; i < end; i++ {
			commits = append(commits, map[string]any{
				"sha":    fmt.Sprintf("sha%d", i),
				"commit": map[string]string{"message": fmt.Sprintf("msg %d", i)},
			})
		}
		_ = json.NewEncoder(w).Encode(commits)
	}))
	defer srv.Close()

	commits, err := Client{BaseURL: srv.URL, HTTPClient: srv.Client()}.ListPullCommits("o/r", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != total {
		t.Fatalf("got %d commits, want %d", len(commits), total)
	}
	if commits[249].SHA != "sha249" || commits[249].Commit.Message != "msg 249" {
		t.Errorf("last commit = %+v", commits[249])
	}
}

func TestPullCallsRateLimitHint(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantHint bool
	}{
		{name: "primary rate limit", status: 403, body: `{"message":"API rate limit exceeded for 192.0.2.1."}`, wantHint: true},
		{name: "secondary rate limit", status: 403, body: `{"message":"You have exceeded a secondary rate limit."}`, wantHint: true},
		{name: "too many requests", status: 429, body: `{"message":"slow down"}`, wantHint: true},
		{name: "plain forbidden", status: 403, body: `{"message":"Resource not accessible by integration"}`},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			c := Client{BaseURL: srv.URL, HTTPClient: srv.Client()}

			_, headErr := c.GetPullHead("o/r", 7)
			_, listErr := c.ListPullCommits("o/r", 7)
			for _, err := range []error{headErr, listErr} {
				if err == nil {
					t.Fatal("expected an error")
				}
				if got := strings.Contains(err.Error(), "set GH_TOKEN"); got != tt.wantHint {
					t.Errorf("hint present = %v, want %v: %v", got, tt.wantHint, err)
				}
				var se *StatusError
				if !errors.As(err, &se) || se.Code != tt.status {
					t.Errorf("error does not wrap the %d StatusError: %v", tt.status, err)
				}
			}
		})
	}
}
