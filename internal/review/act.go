package review

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/partio-io/cli/internal/github"
)

// ReviewedLabel marks an issue the sweep has acted on, open or closed,
// so the sweep never selects it again.
const ReviewedLabel = "minion-reviewed"

// buildLabels describe approvals and runs of the text before the
// review. A keep removes them: the reviewed text starts clean.
var buildLabels = []string{"minion-approved", "minion-failed", "minion-executing", "do-not-build"}

// olderBuildBranch is the head branch an older build of issue pushed.
func olderBuildBranch(issue int) string {
	return fmt.Sprintf("minion/implement-implement-%d", issue)
}

// refusal is the cause of a valid verdict the gate cannot act on. Run
// records it as no verdict; every other error from actable is a failure
// to talk to GitHub.
type refusal string

func (r refusal) Error() string { return string(r) }

// actable checks, outside dry-run, that the gate can act on v for
// issue: an open issue and not a pull request, and for a duplicate
// close a kept issue that is open or that the review closed, which it
// returns. A facts-only check never closes, so it needs no kept issue.
func actable(gh github.Client, repo string, issue github.Issue, v Verdict, factsOnly bool) (*github.Issue, error) {
	switch {
	case issue.PullRequest != nil:
		return nil, refusal(fmt.Sprintf("#%d is a pull request: the gate never changes one", issue.Number))
	case issue.State != "open":
		return nil, refusal(fmt.Sprintf("#%d is %s: the gate acts on open issues only", issue.Number, issue.State))
	case factsOnly || v.Outcome != OutcomeClose || v.CloseReason != ReasonDuplicate:
		return nil, nil
	}
	kept, err := gh.GetIssue(repo, v.DuplicateOf)
	var status *github.StatusError
	if errors.As(err, &status) && status.Code == http.StatusNotFound {
		return nil, refusal(fmt.Sprintf("kept issue #%d does not exist", v.DuplicateOf))
	}
	if err != nil {
		return nil, fmt.Errorf("read kept issue #%d: %w", v.DuplicateOf, err)
	}
	if kept.PullRequest != nil {
		return nil, refusal(fmt.Sprintf("kept issue #%d is a pull request", v.DuplicateOf))
	}
	if kept.State != "open" && !closedByReview(kept) {
		return nil, refusal(fmt.Sprintf("kept issue #%d is closed, but not by a review close, so it holds no verdict for the idea", v.DuplicateOf))
	}
	return &kept, nil
}

// closedByReview reports whether a review close closed issue: closed as
// not planned, with minion-reviewed. Its idea already has a verdict, so
// a later issue with the same idea closes as its duplicate. A review
// close for a built idea is completed, and makes a later issue built,
// not a duplicate.
func closedByReview(issue github.Issue) bool {
	return issue.State == "closed" && issue.StateReason == github.StateReasonNotPlanned && issue.HasLabel(ReviewedLabel)
}

// act applies a verdict to the reviewed issue: one evidence comment,
// for a rewrite the new title and body, the label changes and, for a
// close, the close with its state reason. A rewrite takes the label
// changes of a keep. kept is the issue that a duplicate close names.
// It never changes a pull request; it only names an open one from an
// older build in the comment. The evidence comment is found by its
// marker and updated in place, so a run that failed after the comment
// does not post a second one when it runs again.
func act(gh github.Client, repo string, issue github.Issue, v Verdict, kept *github.Issue) error {
	pulls, err := gh.OpenPullsByHead(repo, olderBuildBranch(issue.Number))
	if err != nil {
		return fmt.Errorf("find older build of #%d: %w", issue.Number, err)
	}
	if err := gh.EnsureLabel(repo, ReviewedLabel, "0e8a16", "The proposal review has acted on this issue"); err != nil {
		return err
	}
	if err := gh.UpsertComment(repo, issue.Number, ReviewMarker, evidenceComment(v, kept, pulls)); err != nil {
		return fmt.Errorf("comment on #%d: %w", issue.Number, err)
	}
	if v.Outcome == OutcomeRewrite {
		if err := gh.EditIssue(repo, issue.Number, oneLine(v.Rewrite.Title), v.Rewrite.Body); err != nil {
			return fmt.Errorf("rewrite #%d: %w", issue.Number, err)
		}
	}
	if v.Outcome != OutcomeClose {
		for _, l := range buildLabels {
			if !issue.HasLabel(l) {
				continue
			}
			if err := gh.RemoveLabel(repo, issue.Number, l); err != nil {
				return fmt.Errorf("label #%d: %w", issue.Number, err)
			}
		}
	}
	if err := gh.AddLabels(repo, issue.Number, ReviewedLabel); err != nil {
		return fmt.Errorf("label #%d: %w", issue.Number, err)
	}
	if v.Outcome != OutcomeClose {
		return nil
	}
	reason := github.StateReasonNotPlanned
	if v.CloseReason == ReasonBuilt {
		reason = github.StateReasonCompleted
	}
	if err := gh.CloseIssue(repo, issue.Number, reason); err != nil {
		return fmt.Errorf("close #%d: %w", issue.Number, err)
	}
	return nil
}
