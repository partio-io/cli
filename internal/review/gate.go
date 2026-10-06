package review

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/partio-io/cli/internal/github"
)

// Config is one gate run for one reviewed issue.
type Config struct {
	VerdictPath string
	Repo        string // owner/name
	Issue       int    // the reviewed issue
	Night       string // UTC date, YYYY-MM-DD, that names the night comment
	DryRun      bool   // record only; change nothing on the reviewed issue
	APIBaseURL  string
	Token       string
	HTTPClient  *http.Client // nil means http.DefaultClient
}

// Result is what the gate found in the verdict file.
type Result struct {
	Valid   bool
	Outcome string // the verdict outcome when Valid
	Cause   string // why there is no verdict when not Valid
}

func (c Config) client() github.Client {
	return github.Client{BaseURL: c.APIBaseURL, Token: c.Token, HTTPClient: c.HTTPClient}
}

// Run loads the verdict for cfg.Issue, acts on the reviewed issue, and
// records the verdict as one row in the tracking issue. A missing or
// invalid verdict, or a rewrite whose new text fails a shape check, is
// "no verdict": Run records that row with its cause and returns a
// Result that is not Valid, with a nil error. A non-nil error means the
// gate could not talk to GitHub. In dry-run mode Run reads the reviewed
// issue and changes nothing on it, and the row of a rewrite that fails
// a shape check also carries the proposed text. Outside dry-run a
// verdict the gate cannot act on is no verdict: a closed issue, a pull
// request, or a duplicate close whose kept issue is not open.
func Run(cfg Config) (Result, error) {
	gh := cfg.client()
	issue, err := gh.GetIssue(cfg.Repo, cfg.Issue)
	if err != nil {
		return Result{}, fmt.Errorf("read issue #%d: %w", cfg.Issue, err)
	}

	res := Result{}
	v, err := LoadVerdict(cfg.VerdictPath, cfg.Issue)
	badShape := false
	if err == nil && v.Outcome == OutcomeRewrite {
		err = checkShape(issue.Body, *v.Rewrite)
		badShape = err != nil
	}
	var kept *github.Issue
	if err == nil && !cfg.DryRun {
		kept, err = actable(gh, cfg.Repo, issue, v)
		var no refusal
		if err != nil && !errors.As(err, &no) {
			return Result{}, err
		}
	}
	var r string
	if err != nil {
		res.Cause = err.Error()
		r = noVerdictRow(issue, res.Cause)
		if badShape && cfg.DryRun {
			r += indent(rewriteDetails(*v.Rewrite))
		}
	} else {
		res.Valid, res.Outcome = true, v.Outcome
		r = row(issue, v, cfg.DryRun)
	}
	if res.Valid && !cfg.DryRun {
		if err := act(gh, cfg.Repo, issue, v, kept); err != nil {
			return Result{}, err
		}
	}

	tracking, err := findOrCreateTracking(gh, cfg.Repo)
	if err != nil {
		return Result{}, err
	}
	nights, err := appendRow(gh, cfg.Repo, tracking.Number, cfg.Night, r)
	if err != nil {
		return Result{}, err
	}
	if err := writeSummary(gh, cfg.Repo, tracking, nights); err != nil {
		return Result{}, err
	}
	return res, nil
}
