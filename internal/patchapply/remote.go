package patchapply

import (
	"fmt"
	"strings"
)

// pushRemote is the target the round fetches from and pushes to. An
// explicit remote wins, which is what lets a test drive the whole
// round against a bare repository on disk.
func pushRemote(cfg Config) string {
	if cfg.Remote != "" {
		return cfg.Remote
	}
	return remoteURL(cfg.Repo, cfg.Token)
}

// remoteURL is the HTTPS URL for repo, carrying token.
//
// The token must be the personal access token. A push made with the
// default Actions token triggers no follow-up workflow run, and the
// next repair round is a follow-up run: the fix would land, the loop
// would stop, and nothing would say so.
func remoteURL(repo, token string) string {
	return fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", token, repo)
}

// redact removes the token from text that reaches a job log.
func redact(text, token string) string {
	if token == "" {
		return text
	}
	return strings.ReplaceAll(text, token, "***")
}
