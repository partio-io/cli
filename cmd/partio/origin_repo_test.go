package main

import "testing"

func TestRepoFromRemoteURL(t *testing.T) {
	tests := []struct {
		url    string
		want   string
		wantOK bool
	}{
		{url: "https://github.com/owner/repo.git", want: "owner/repo", wantOK: true},
		{url: "https://github.com/owner/repo", want: "owner/repo", wantOK: true},
		{url: "https://github.com/owner/repo/\n", want: "owner/repo", wantOK: true},
		{url: "https://user:tok@ghe.example.com/owner/repo.git", want: "owner/repo", wantOK: true},
		{url: "git@github.com:owner/repo.git", want: "owner/repo", wantOK: true},
		{url: "github.com:owner/repo", want: "owner/repo", wantOK: true},
		{url: "ssh://git@github.com:22/owner/repo.git", want: "owner/repo", wantOK: true},
		{url: "git://github.com/owner/repo.git", want: "owner/repo", wantOK: true},
		{url: "file:///tmp/owner/repo"},
		{url: "/tmp/owner/repo"},
		{url: "../repo"},
		{url: "https://github.com/repo"},
		{url: "git@github.com:repo.git"},
		{url: ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, ok := repoFromRemoteURL(tt.url)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("repoFromRemoteURL(%q) = %q, %v; want %q, %v", tt.url, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
