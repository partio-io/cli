package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDo(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		out     bool
		wantErr string
		wantID  int
	}{
		{name: "decodes a 2xx body", status: 200, body: `{"id":7}`, out: true, wantID: 7},
		{name: "ignores the body when out is nil", status: 204, body: "not json"},
		{
			name: "rejects a non-2xx response", status: 422, body: `{"message":"bad"}`, out: true,
			wantErr: `github: GET /repos/o/r/pulls/1: 422 Unprocessable Entity: {"message":"bad"}`,
		},
		{
			name: "keeps 512 bytes of the error body", status: 500, body: strings.Repeat("x", 600),
			wantErr: "github: GET /repos/o/r/pulls/1: 500 Internal Server Error: " + strings.Repeat("x", 512),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var accept, auth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				accept, auth = r.Header.Get("Accept"), r.Header.Get("Authorization")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := Client{BaseURL: srv.URL, Token: "tok", HTTPClient: srv.Client()}
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/repos/o/r/pulls/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				ID int `json:"id"`
			}
			var out any
			if tt.out {
				out = &got
			}
			err = c.Do(req, out)

			if accept != "application/vnd.github+json" {
				t.Errorf("Accept = %q", accept)
			}
			if auth != "Bearer tok" {
				t.Errorf("Authorization = %q", auth)
			}
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.wantID {
				t.Errorf("id = %d, want %d", got.ID, tt.wantID)
			}
		})
	}
}

func TestDoUsesDefaultClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Client{Token: "tok"}).Do(req, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
