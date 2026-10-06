package github

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// EnsureLabel creates the label only on a 404. Any other failure of the
// read is an error, never a reason to create.
func TestEnsureLabel(t *testing.T) {
	tests := []struct {
		name      string
		getStatus int
		wantErr   bool
		want      []string
	}{
		{name: "present", getStatus: http.StatusOK, want: []string{"GET /repos/o/r/labels/a b"}},
		{name: "absent", getStatus: http.StatusNotFound, want: []string{"GET /repos/o/r/labels/a b", "POST /repos/o/r/labels"}},
		{name: "broken", getStatus: http.StatusInternalServerError, wantErr: true, want: []string{"GET /repos/o/r/labels/a b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, r.Method+" "+r.URL.Path)
				if r.Method == http.MethodGet {
					w.WriteHeader(tt.getStatus)
					return
				}
				w.WriteHeader(http.StatusCreated)
			}))
			defer srv.Close()

			err := Client{BaseURL: srv.URL}.EnsureLabel("o/r", "a b", "ffffff", "d")

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("requests = %q, want %q", got, tt.want)
			}
		})
	}
}

// RemoveLabel treats GitHub's 404 for a label the issue does not carry
// as done. Any other failure is an error.
func TestRemoveLabel(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "present", status: http.StatusOK},
		{name: "absent", status: http.StatusNotFound},
		{name: "broken", status: http.StatusInternalServerError, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, r.Method+" "+r.URL.Path)
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			err := Client{BaseURL: srv.URL}.RemoveLabel("o/r", 12, "do-not-build")

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if want := []string{"DELETE /repos/o/r/issues/12/labels/do-not-build"}; !slices.Equal(got, want) {
				t.Errorf("requests = %q, want %q", got, want)
			}
		})
	}
}
