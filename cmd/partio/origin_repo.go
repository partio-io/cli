package main

import (
	"net/url"
	"os/exec"
	"strings"
)

// originRepo returns the owner/repo that the origin remote of the
// repository in dir points at. ok is false when there is no origin or its
// URL names no owner/repo.
func originRepo(dir string) (repo string, ok bool) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return repoFromRemoteURL(string(out))
}

// repoFromRemoteURL extracts owner/repo from a remote URL: a URL with a
// scheme and host (https://, ssh://, git://) or the scp-like
// [user@]host:owner/repo form. A trailing ".git" is dropped. Local paths and
// file:// URLs name no owner/repo.
func repoFromRemoteURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	var path string
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return "", false
		}
		path = u.Path
	} else {
		host, rest, found := strings.Cut(raw, ":")
		if !found || host == "" || strings.Contains(host, "/") {
			return "", false
		}
		path = rest
	}

	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	segs := strings.Split(path, "/")
	if len(segs) < 2 {
		return "", false
	}
	owner, repo := segs[len(segs)-2], segs[len(segs)-1]
	if !prNamePattern.MatchString(owner) || !prNamePattern.MatchString(repo) {
		return "", false
	}
	return owner + "/" + repo, true
}
