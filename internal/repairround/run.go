// Package repairround decides whether a pull request may receive
// another automated repair round for one check, and which round it
// would be.
//
// The decision is accounting, not prohibition: repair commits carry a
// marker naming the check that produced them, and a run counts the
// markers already on the branch. Under the cap the run proceeds and
// knows its round number; at the cap it declines. Each check keeps its
// own budget.
package repairround

import "net/http"

// Config carries everything one accounting run needs.
type Config struct {
	Repo       string       // owner/name
	PRNumber   int          // pull request whose branch is counted
	Check      string       // check whose budget is in question, e.g. "dead-code"
	MaxRounds  int          // cap for this check; zero means DefaultMaxRounds
	APIBaseURL string       // GitHub API root, e.g. https://api.github.com
	Token      string       // token for the commits read
	HTTPClient *http.Client // nil means http.DefaultClient
}

// RunEligibility reads the pull request's labels and head repository
// and reports whether a repair round may push to it at all.
//
// This is the question that comes before the count. An error means the
// pull request could not be read, and the caller must decline rather
// than assume: a wrong "yes" here is a commit on someone's branch.
func RunEligibility(cfg Config) (Eligibility, error) {
	labels, headRepo, err := prMeta(cfg)
	if err != nil {
		return Eligibility{}, err
	}
	return Eligible(labels, headRepo, cfg.Repo), nil
}

// Run reads the pull request branch's commit subjects and reports
// whether another repair round for cfg.Check may start.
func Run(cfg Config) (Decision, error) {
	subjects, err := prSubjects(cfg)
	if err != nil {
		return Decision{}, err
	}
	maxRounds := cfg.MaxRounds
	if maxRounds <= 0 {
		maxRounds = DefaultMaxRounds
	}
	return Decide(subjects, cfg.Check, maxRounds), nil
}
