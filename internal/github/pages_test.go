package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetPages(t *testing.T) {
	tests := []struct {
		name      string
		sizes     []int // items served per page
		path      string
		maxPages  int
		stopAt    int // visit returns false on this page; 0 means never
		wantPages int
		wantItems int
	}{
		{name: "one short page", sizes: []int{3}, path: "/items", wantPages: 1, wantItems: 3},
		{name: "stops on a short page", sizes: []int{100, 100, 5, 100}, path: "/items", wantPages: 3, wantItems: 205},
		{name: "stops on an empty page", sizes: []int{100, 0}, path: "/items", wantPages: 2, wantItems: 100},
		{name: "stops at maxPages", sizes: []int{100, 100, 100}, path: "/items", maxPages: 2, wantPages: 2, wantItems: 200},
		{name: "stops when visit says so", sizes: []int{100, 100, 100}, path: "/items", stopAt: 2, wantPages: 2, wantItems: 200},
		{name: "keeps an existing query", sizes: []int{1}, path: "/items?state=all", wantPages: 1, wantItems: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				queries = append(queries, r.URL.RawQuery)
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				items := make([]int, tt.sizes[page-1])
				_ = json.NewEncoder(w).Encode(items)
			}))
			defer srv.Close()

			c := Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
			var pages, items int
			err := GetPages(c, tt.path, tt.maxPages, func(page []int) bool {
				pages++
				items += len(page)
				return pages != tt.stopAt
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pages != tt.wantPages || len(queries) != tt.wantPages {
				t.Errorf("pages = %d, requests = %d, want %d", pages, len(queries), tt.wantPages)
			}
			if items != tt.wantItems {
				t.Errorf("items = %d, want %d", items, tt.wantItems)
			}
			wantFirst := "per_page=100&page=1"
			if tt.path == "/items?state=all" {
				wantFirst = "state=all&" + wantFirst
			}
			if queries[0] != wantFirst {
				t.Errorf("first query = %q, want %q", queries[0], wantFirst)
			}
		})
	}
}

func TestGetPagesReturnsTheRequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer srv.Close()

	err := GetPages(Client{BaseURL: srv.URL, HTTPClient: srv.Client()}, "/items", 0, func([]int) bool { return true })
	want := "github: GET /items: 404 Not Found: gone\n"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}
