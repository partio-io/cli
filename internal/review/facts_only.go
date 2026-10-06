package review

import (
	"fmt"
	"strings"
	"time"

	"github.com/partio-io/cli/internal/github"
	"github.com/partio-io/cli/internal/premise"
)

// closeCommentPrefix starts the evidence comment of a close.
const closeCommentPrefix = ReviewMarker + "\n## Proposal review: " + OutcomeClose + "\n"

// checksFactsOnly reports whether the gate checks only the facts of
// issue, and then never closes, rewrites or relabels it. That is so when
// the operator wrote the issue, as it carries no minion-proposal label,
// or reopened it after the review closed it: a reopen is the operator's
// overrule. The runner posts as the operator, so the actor of the
// reopen tells nothing, but the review itself never reopens an issue.
// The review's close comment is the one the gate updated last before it
// closed the issue, and the gate never touches a closed issue, so any
// reopen at or after the comment's last edit follows that close.
func checksFactsOnly(gh github.Client, repo string, issue github.Issue) (bool, error) {
	if !issue.HasLabel(proposalLabel) {
		return true, nil
	}
	events, err := gh.Timeline(repo, issue.Number)
	if err != nil {
		return false, err
	}
	var closedAt *time.Time
	for _, ev := range events {
		switch {
		case ev.Event == "commented" && strings.HasPrefix(ev.Body, closeCommentPrefix):
			closedAt = &ev.UpdatedAt
		case ev.Event == "reopened" && closedAt != nil && !ev.CreatedAt.Before(*closedAt):
			return true, nil
		}
	}
	return false, nil
}

// premiseBlocks reports whether the premise of v stops a facts-only
// issue: a claim that fails, or one that nothing settled, is not a pass.
// Each claim counts, not only the premise verdict: the verdict check
// does not tie the two together, so a "holds" premise can carry an
// unresolved claim.
func premiseBlocks(v Verdict) bool {
	p := v.Premise
	return p.Verdict == Fails || p.Verdict == Unresolved || p.hasClaim(Fails) || p.hasClaim(Unresolved)
}

// actFactsOnly applies only the premise of v to issue and ignores its
// outcome. A premise that blocks removes do-not-build and then adds it,
// so the label event fires again on every block, as the stage gate
// does. Both a block and a pass post the marked comment with every
// claim, so a pass also leaves its evidence. Nothing else on the issue
// changes.
func actFactsOnly(gh github.Client, repo string, issue github.Issue, v Verdict) error {
	if premiseBlocks(v) {
		if err := gh.RemoveLabel(repo, issue.Number, premise.BlockingLabel); err != nil {
			return fmt.Errorf("label #%d: %w", issue.Number, err)
		}
		if err := gh.AddLabels(repo, issue.Number, premise.BlockingLabel); err != nil {
			return fmt.Errorf("label #%d: %w", issue.Number, err)
		}
	}
	if err := gh.CreateComment(repo, issue.Number, premiseComment(v)); err != nil {
		return fmt.Errorf("comment on #%d: %w", issue.Number, err)
	}
	return nil
}

// premiseComment is the facts-only account of one run: what it decided
// for the build, and every claim with its evidence, verdict and excerpt.
// When the excerpts push it past GitHub's size limit, it leaves them out
// and says so.
func premiseComment(v Verdict) string {
	if body := renderPremise(v, true); len(body) <= maxCommentBody {
		return body
	}
	return renderPremise(v, false)
}

func renderPremise(v Verdict, excerpts bool) string {
	var b strings.Builder
	b.WriteString(premise.GateCommentMarker + "\n")
	fmt.Fprintf(&b, "## Premise check: %s\n\n", v.Premise.Verdict)
	if premiseBlocks(v) {
		fmt.Fprintf(&b, "The build stops, and the issue carries `%s`. "+
			"Correct the issue text, or remove the label and run the build again.\n\n", premise.BlockingLabel)
	} else {
		b.WriteString("The build proceeds.\n\n")
	}
	b.WriteString("The review checks facts only on this issue: it does not close, rewrite or relabel it.\n\n")
	writeClaims(&b, v.Premise.Claims, excerpts)
	return b.String()
}
