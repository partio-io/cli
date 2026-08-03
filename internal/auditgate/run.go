// Package auditgate turns a minion audit verdict file into a CI
// outcome: green only for a valid pass verdict, red for everything
// else (fail-closed). On red it maintains a single PR comment
// describing the findings.
package auditgate

import (
	"fmt"
	"net/http"
)

// Config carries everything one gate run needs.
type Config struct {
	VerdictPath string       // path to the verdict JSON file
	Repo        string       // owner/name
	PRNumber    int          // pull request to comment on
	AuditName   string       // names this audit in the comment, e.g. "dead-code"
	APIBaseURL  string       // GitHub API root, e.g. https://api.github.com
	Token       string       // token for the comment upsert
	HTTPClient  *http.Client // nil means http.DefaultClient
}

// Result is the gate's outcome. Green maps to exit 0, red to exit 1.
type Result struct {
	Green  bool
	Reason string
}

// Run maps the verdict file to a check outcome and maintains the PR
// comment on red.
func Run(cfg Config) (Result, error) {
	v, err := loadVerdict(cfg.VerdictPath)
	if err != nil {
		if upErr := upsertComment(cfg, failClosedBody(cfg.AuditName, err)); upErr != nil {
			return Result{}, upErr
		}
		return Result{Green: false, Reason: fmt.Sprintf("fail-closed: %v", err)}, nil
	}
	if v.Status == "pass" {
		return Result{Green: true, Reason: "verdict: pass"}, nil
	}
	if err := upsertComment(cfg, failBody(cfg.AuditName, v)); err != nil {
		return Result{}, err
	}
	return Result{Green: false, Reason: fmt.Sprintf("verdict: fail with %d finding(s)", len(v.Findings))}, nil
}
