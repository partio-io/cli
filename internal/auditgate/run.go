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

	// Round is the repair round this run belongs to, as counted by
	// internal/repairround; 1 is the first attempt. MaxRounds is the
	// cap for this check. Both zero means the caller keeps no round
	// accounting, and every failure reads as a first-pass failure.
	Round     int
	MaxRounds int
}

// budgetSpent reports whether this run is the last repair round the
// check gets. The gate runs after the round's repair attempt, so a
// surviving finding at the cap is one no further round will address.
func (c Config) budgetSpent() bool {
	return c.MaxRounds > 0 && c.Round >= c.MaxRounds
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
	if cfg.budgetSpent() {
		if err := upsertComment(cfg, exhaustedBody(cfg.AuditName, v, cfg.Round)); err != nil {
			return Result{}, err
		}
		return Result{Green: false, Reason: fmt.Sprintf(
			"verdict: fail with %d finding(s) after %d repair round(s); budget spent",
			len(v.Findings), cfg.Round)}, nil
	}
	if err := upsertComment(cfg, failBody(cfg.AuditName, v)); err != nil {
		return Result{}, err
	}
	return Result{Green: false, Reason: fmt.Sprintf("verdict: fail with %d finding(s)", len(v.Findings))}, nil
}
