package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// fakeComments serves the issue-comments endpoints for repo o/r,
// issue 9, and records every write.
type fakeComments struct {
	comments []comment
	writes   []string // "METHOD path body"
}

func (f *fakeComments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/issues/9/comments":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		start := min((page-1)*PerPage, len(f.comments))
		end := min(start+PerPage, len(f.comments))
		_ = json.NewEncoder(w).Encode(f.comments[start:end])
	case r.Method == http.MethodPost || r.Method == http.MethodPatch:
		var payload struct {
			Body string `json:"body"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		f.writes = append(f.writes, fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, payload.Body))
	default:
		http.NotFound(w, r)
	}
}

func TestUpsertComment(t *testing.T) {
	filler := func(n int) []comment {
		cs := make([]comment, n)
		for i := range cs {
			cs[i] = comment{ID: int64(i + 1), Body: "unrelated"}
		}
		return cs
	}
	tests := []struct {
		name      string
		comments  []comment
		wantWrite string
	}{
		{
			name:      "creates when no comment matches",
			comments:  filler(3),
			wantWrite: "POST /repos/o/r/issues/9/comments Minion audit — x new",
		},
		{
			name:      "updates the first match in place",
			comments:  append(filler(2), comment{ID: 50, Body: "Minion audit — x old"}, comment{ID: 51, Body: "Minion audit — x older"}),
			wantWrite: "PATCH /repos/o/r/issues/comments/50 Minion audit — x new",
		},
		{
			name:      "finds a match past the first page",
			comments:  append(filler(150), comment{ID: 500, Body: "Minion audit — x old"}),
			wantWrite: "PATCH /repos/o/r/issues/comments/500 Minion audit — x new",
		},
		{
			name:      "does not match a prefix in the middle of a body",
			comments:  []comment{{ID: 1, Body: "see Minion audit — x"}},
			wantWrite: "POST /repos/o/r/issues/9/comments Minion audit — x new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeComments{comments: tt.comments}
			srv := httptest.NewServer(fake)
			defer srv.Close()

			c := Client{BaseURL: srv.URL, Token: "tok", HTTPClient: srv.Client()}
			if err := c.UpsertComment("o/r", 9, "Minion audit — x", "Minion audit — x new"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fake.writes) != 1 || fake.writes[0] != tt.wantWrite {
				t.Errorf("writes = %q, want [%q]", fake.writes, tt.wantWrite)
			}
		})
	}
}

func TestUpsertCommentErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{
			name: "list fails",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "nope", http.StatusForbidden)
			},
			want: "list comments: github: GET /repos/o/r/issues/9/comments: 403 Forbidden: nope\n",
		},
		{
			name: "create fails",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte("[]"))
					return
				}
				http.Error(w, "nope", http.StatusForbidden)
			},
			want: "send comment: github: POST /repos/o/r/issues/9/comments: 403 Forbidden: nope\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			c := Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
			err := c.UpsertComment("o/r", 9, "p", "p body")
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}
