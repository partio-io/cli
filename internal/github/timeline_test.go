package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// Timeline follows every page and keeps each event's kind, comment body
// and times, so a caller can order a reopen against a comment.
func TestTimeline(t *testing.T) {
	events := make([]map[string]any, 0, PerPage+2)
	for i := range PerPage {
		events = append(events, map[string]any{"event": "labeled", "created_at": "2026-10-01T00:00:00Z", "label": map[string]any{"name": fmt.Sprint(i)}})
	}
	events = append(events,
		map[string]any{"event": "commented", "id": 7, "body": "hello", "created_at": "2026-10-02T10:00:00Z", "updated_at": "2026-10-03T10:00:00Z"},
		map[string]any{"event": "reopened", "created_at": "2026-10-04T10:00:00Z"},
	)
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path != "/repos/o/r/issues/9/timeline" {
			http.NotFound(w, r)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		start := min((page-1)*PerPage, len(events))
		end := min(start+PerPage, len(events))
		_ = json.NewEncoder(w).Encode(events[start:end])
	}))
	defer srv.Close()

	got, err := Client{BaseURL: srv.URL}.Timeline("o/r", 9)
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	if len(got) != PerPage+2 || len(paths) != 2 {
		t.Fatalf("got %d events in %d requests, want %d in 2", len(got), len(paths), PerPage+2)
	}
	at := func(s string) time.Time {
		tm, _ := time.Parse(time.RFC3339, s)
		return tm
	}
	want := []TimelineEvent{
		{Event: "commented", ID: 7, Body: "hello", CreatedAt: at("2026-10-02T10:00:00Z"), UpdatedAt: at("2026-10-03T10:00:00Z")},
		{Event: "reopened", CreatedAt: at("2026-10-04T10:00:00Z")},
	}
	for i, w := range want {
		if g := got[PerPage+i]; g != w {
			t.Errorf("event %d = %+v, want %+v", PerPage+i, g, w)
		}
	}
}
