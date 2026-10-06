package review

import (
	"fmt"
	"strings"

	"github.com/partio-io/cli/internal/github"
)

// ReviewMarker opens every evidence comment the gate posts.
const ReviewMarker = "<!-- partio:review:v1 -->"

// evidenceComment is the comment the gate posts on the reviewed issue:
// the outcome and its reason, for a rewrite what changed, every claim
// with its evidence, verdict, excerpt and any correction, the fit, built
// and duplicate decisions, and any open pull request from an older
// build. kept is the issue that stays for a duplicate close, and nil
// otherwise.
// When the excerpts push the comment past GitHub's size limit, the
// comment leaves them out and says so.
func evidenceComment(v Verdict, kept *github.Issue, pulls []github.PullRequest) string {
	if body := renderEvidence(v, kept, pulls, true); len(body) <= maxCommentBody {
		return body
	}
	return renderEvidence(v, kept, pulls, false)
}

func renderEvidence(v Verdict, kept *github.Issue, pulls []github.PullRequest, excerpts bool) string {
	var b strings.Builder
	b.WriteString(ReviewMarker + "\n")
	fmt.Fprintf(&b, "## Proposal review: %s\n\n", v.Outcome)
	fmt.Fprintf(&b, "**Outcome:** %s · **Reason:** %s\n\n", v.Outcome, reason(v))
	if kept != nil {
		fmt.Fprintf(&b, "Duplicate of #%d, which stays: [%s](%s)\n\n", kept.Number, oneLine(kept.Title), kept.HTMLURL)
	}
	if v.Outcome == OutcomeRewrite {
		b.WriteString("### What changed\n\n")
		if len(v.Rewrite.Changes) == 0 {
			b.WriteString("The review lists no change.\n")
		}
		for _, c := range v.Rewrite.Changes {
			fmt.Fprintf(&b, "- %s\n", oneLine(c))
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "### Premise: %s\n\n", v.Premise.Verdict)
	if v.Outcome == OutcomeClose && v.CloseReason == ReasonFalsePremise {
		b.WriteString("To keep the idea, reopen the issue and correct its text. The correction " +
			"under each failing claim is the fact that holds instead.\n\n")
	}
	writeClaims(&b, v.Premise.Claims, excerpts)
	b.WriteString("\n### Decisions\n\n")
	fmt.Fprintf(&b, "- **Fit:** %s · %s\n", yesNo(v.Fit.Applies, "applies", "does not apply"), oneLine(v.Fit.Reason))
	fmt.Fprintf(&b, "- **Built:** %s", yesNo(v.Built.Built, "yes", "no"))
	if v.Built.Evidence != "" {
		fmt.Fprintf(&b, " · %s", oneLine(v.Built.Evidence))
	}
	b.WriteString("\n- **Duplicates:**")
	if len(v.Duplicates) == 0 {
		b.WriteString(" none found\n")
	} else {
		b.WriteString("\n")
		for _, d := range v.Duplicates {
			fmt.Fprintf(&b, "  - #%d · %s · %s\n", d.Issue, yesNo(d.Same, "same idea", "different idea"), oneLine(d.Why))
		}
	}
	if len(pulls) > 0 {
		b.WriteString("\n### Open pull request from an older build\n\n")
		for _, p := range pulls {
			fmt.Fprintf(&b, "- #%d · [%s](%s) · the review does not change it\n", p.Number, oneLine(p.Title), p.HTMLURL)
		}
	}
	return b.String()
}

// writeClaims writes every claim with its verdict, its correction when it
// has one, its evidence and, when excerpts is true, its excerpt.
func writeClaims(b *strings.Builder, claims []Claim, excerpts bool) {
	if len(claims) == 0 {
		b.WriteString("No checkable claims.\n\n")
	}
	if !excerpts {
		b.WriteString("Excerpts left out: with them the comment passes GitHub's size limit.\n\n")
	}
	for _, c := range claims {
		fmt.Fprintf(b, "- **%s** · %s\n", c.Verdict, oneLine(c.Claim))
		if c.Correction != "" {
			fmt.Fprintf(b, "  - Correction: %s\n", oneLine(c.Correction))
		}
		fmt.Fprintf(b, "  - Evidence: %s\n", oneLine(c.Evidence))
		if excerpts && c.Excerpt != "" {
			fence := fenceFor(c.Excerpt)
			b.WriteString("  - Excerpt:\n\n")
			b.WriteString(indent(indent(fence + "\n" + strings.TrimRight(c.Excerpt, "\n") + "\n" + fence)))
			b.WriteString("\n")
		}
	}
}

func yesNo(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}
