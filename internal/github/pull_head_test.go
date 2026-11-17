package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetPullHead(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantHead *PullRepo
		wantSHA  string
		wantBase string
	}{
		{
			name:     "decodes head and base",
			body:     `{"head":{"sha":"abc","repo":{"full_name":"f/r","clone_url":"https://github.com/f/r.git"}},"base":{"repo":{"full_name":"o/r"}}}`,
			wantHead: &PullRepo{FullName: "f/r", CloneURL: "https://github.com/f/r.git"},
			wantSHA:  "abc",
			wantBase: "o/r",
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
