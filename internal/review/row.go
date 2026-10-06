package review

import (
	"fmt"
	"strings"

	"github.com/partio-io/cli/internal/github"
)

// row is one tracking line for a valid verdict: number, title, verdict,
// reason and link. A rewrite carries its full proposed text in a
// collapsed section, indented so it stays under its row, with the body
// fenced so its own markup cannot close the section or hide later rows.
func row(issue github.Issue, v Verdict, dryRun bool) string {
	verdict := "**" + v.Outcome + "**"
	if dryRun {
		verdict += " (dry run)"
	}
	line := fmt.Sprintf("- #%d · %s · %s · %s · [link](%s)\n",
		issue.Number, oneLine(issue.Title), verdict, reason(v), issue.HTMLURL)
	if v.Outcome != OutcomeRewrite {
		return line
	}
	return line + indent(rewriteDetails(*v.Rewrite))
}

// noVerdictRow is the tracking line for a missing or invalid verdict.
func noVerdictRow(issue github.Issue, cause string) string {
	return fmt.Sprintf("- #%d · %s · **no verdict** · %s · [link](%s)\n",
		issue.Number, oneLine(issue.Title), oneLine(cause), issue.HTMLURL)
}

func reason(v Verdict) string {
	switch {
	case v.Outcome == OutcomeClose && v.CloseReason == ReasonDuplicate:
		return fmt.Sprintf("duplicate of #%d", v.DuplicateOf)
	case v.Outcome == OutcomeClose:
		return v.CloseReason
	case v.Outcome == OutcomeRewrite && len(v.Rewrite.Changes) > 0:
		return oneLine(strings.Join(v.Rewrite.Changes, "; "))
	}
	return oneLine(v.Fit.Reason)
}

func rewriteDetails(rw Rewrite) string {
	var b strings.Builder
	b.WriteString("\n<details><summary>Proposed rewrite</summary>\n\n")
	fmt.Fprintf(&b, "**Title:** %s\n\n", oneLine(rw.Title))
	fence := fenceFor(rw.Body)
	b.WriteString(fence + "markdown\n")
	b.WriteString(strings.TrimRight(rw.Body, "\n"))
	b.WriteString("\n" + fence + "\n\n</details>\n\n")
	return b.String()
}

// fenceFor returns a backtick fence longer than any backtick run in
// text, so text cannot close it.
func fenceFor(text string) string {
	longest, run := 0, 0
	for _, r := range text {
		if r == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	return strings.Repeat("`", max(3, longest+1))
}

// indent nests text under a list item, so lines that start with "- "
// in a rewrite body never read as rows of their own.
func indent(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = "  " + l
		}
	}
	return strings.Join(lines, "\n")
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
