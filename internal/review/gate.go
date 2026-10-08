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
	Build       bool   // a build reviews its issue: act, and mark the row as a build
	APIBaseURL  string
	Token       string
	HTTPClient  *http.Client // nil means http.DefaultClient
}

// Result is what the gate found in the verdict file.
type Result struct {
	Valid     bool
	Outcome   string // the verdict outcome when Valid
	Cause     string // why there is no verdict when not Valid
	FactsOnly bool   // the gate applied only the premise, outside dry-run
	Blocked   bool   // the build stops: a close, or a facts-only premise that blocks
}

// mark is the suffix of the verdict column: what kind of run wrote the
// row. A sweep that acts writes no mark.
func (c Config) mark() string {
	switch {
	case c.DryRun:
		return dryRunMark
	case c.Build:
		return buildMark
	}
	return ""
}

func (c Config) client() github.Client {
	return github.Client{BaseURL: c.APIBaseURL, Token: c.Token, HTTPClient: c.HTTPClient}
}

// Run loads the verdict for cfg.Issue, acts on the reviewed issue, and
// records the verdict as one row in the tracking issue. A missing or
// invalid verdict, or a rewrite whose new text fails a shape check, is
// "no verdict": Run records that row with its cause and returns a
// Result that is not Valid, with a nil error. A non-nil error means the
// gate could not talk to GitHub; when the tracking issue fails after
// the gate acted, the Result still names the verdict it acted on. In
// dry-run mode Run reads the reviewed issue and changes nothing on it,
// and the row of a rewrite that fails a shape check also carries the
// proposed text. Outside dry-run a
// verdict the gate cannot act on is no verdict: a closed issue, a pull
// request, or a duplicate close whose kept issue is not open. A build
// run acts as a real run does, and its row carries the build mark. On
// an issue the operator wrote, the gate checks facts only: it applies
// the premise of the verdict and ignores its outcome.
func Run(cfg Config) (Result, error) {
	gh := cfg.client()
	issue, err := gh.GetIssue(cfg.Repo, cfg.Issue)
	if err != nil {
		return Result{}, fmt.Errorf("read issue #%d: %w", cfg.Issue, err)
	}

	res := Result{}
	if !cfg.DryRun {
		if res.FactsOnly, err = checksFactsOnly(gh, cfg.Repo, issue); err != nil {
			return Result{}, err
		}
	}
	v, badShape, err := verdictFor(cfg.VerdictPath, cfg.Issue, issue.Body, res.FactsOnly)
	var kept *github.Issue
	if err == nil && !cfg.DryRun {
		kept, err = actable(gh, cfg.Repo, issue, v, res.FactsOnly)
		var no refusal
		if err != nil && !errors.As(err, &no) {
			return Result{}, err
		}
	}
	var r string
	if err != nil {
		res.Cause = err.Error()
		// A dry run's "no verdict" row has always gone unmarked; only
		// a build marks its own.
		mark := ""
		if cfg.Build {
			mark = buildMark
		}
		r = noVerdictRow(issue, res.Cause, mark)
		if badShape && cfg.DryRun {
			r += indent(rewriteDetails(*v.Rewrite))
		}
	} else {
		res.Valid, res.Outcome = true, v.Outcome
		res.Blocked = v.Outcome == OutcomeClose
		if res.FactsOnly {
			res.Blocked = premiseBlocks(v)
		}
		r = row(issue, v, cfg.mark())
	}
	switch {
	case !res.Valid || cfg.DryRun:
	case res.FactsOnly:
		if err := actFactsOnly(gh, cfg.Repo, issue, v); err != nil {
			return Result{}, err
		}
	default:
		if err := act(gh, cfg.Repo, issue, v, kept); err != nil {
			return Result{}, err
		}
	}

	// Any action on the issue is done from here, so an error still
	// returns res: a build can tell a close it made from no verdict.
	tracking, err := findOrCreateTracking(gh, cfg.Repo)
	if err != nil {
		return res, err
	}
	nights, err := appendRow(gh, cfg.Repo, tracking.Number, cfg.Night, r)
	if err != nil {
		return res, err
	}
	if err := writeSummary(gh, cfg.Repo, tracking, nights); err != nil {
		return res, err
	}
	return res, nil
}

// verdictFor loads the verdict for issue from path and, for a rewrite,
// checks the shape of the new body against oldBody. A facts-only check
// never uses the rewrite text, so its shape does not matter there.
// badShape reports a rewrite that failed the shape check. Run and Check
// share it, so the review's own check never drifts from the gate's.
func verdictFor(path string, issue int, oldBody string, factsOnly bool) (v Verdict, badShape bool, err error) {
	v, err = LoadVerdict(path, issue)
	if err != nil || v.Outcome != OutcomeRewrite || factsOnly {
		return v, false, err
	}
	if err := checkShape(oldBody, *v.Rewrite); err != nil {
		return v, true, err
	}
	return v, false, nil
}

// Check runs the checks of the gate on a verdict file, and changes
// nothing. The review program runs it on its own verdict before its
// session ends, so the session can still fix a verdict that the gate
// would record as "no verdict". It returns the cause of "no verdict",
// or "" when the verdict passes. It checks the shape of every rewrite,
// because it does not decide the facts-only mode, and it leaves to the
// gate what it checks on GitHub when it acts: an open issue, and for a
// duplicate a kept issue that is open or that the review closed. A
// non-nil error means that Check could not read the issue.
func Check(cfg Config) (string, error) {
	issue, err := cfg.client().GetIssue(cfg.Repo, cfg.Issue)
	if err != nil {
		return "", fmt.Errorf("read issue #%d: %w", cfg.Issue, err)
	}
	if _, _, err := verdictFor(cfg.VerdictPath, cfg.Issue, issue.Body, false); err != nil {
		return err.Error(), nil
	}
	return "", nil
}
