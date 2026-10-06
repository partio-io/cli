package review

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// noVerdictLimit is the number of "no verdict" rows after which the
// sweep stops picking an issue: the review cannot settle it, so the
// operator must.
const noVerdictLimit = 2

// rowLine matches one row of a night comment, as row and noVerdictRow
// write it: the issue number, the verdict, the dry-run or build mark
// and the reason. The indented details under a rewrite row never match. The
// title match is lazy, so a reason that quotes a bold verdict cannot
// take the place of the real one.
var rowLine = regexp.MustCompile(`^- #(\d+) · .*? · \*\*(keep|rewrite|close|no verdict)\*\*( \((?:dry run|build)\))? · (.*) · \[link\]\(.*\)$`)

// logSummary is what the night comments of the tracking issue add up to.
type logSummary struct {
	verdicts tally // per verdict, "no verdict" included
	reasons  tally // per close reason
	needsYou []int // issues with noVerdictLimit "no verdict" rows since their last acted verdict
}

// tally counts rows per key, and how many of them are dry runs.
type tally map[string]struct{ rows, dryRun int }

func (t tally) add(key string, dryRun bool) {
	c := t[key]
	c.rows++
	if dryRun {
		c.dryRun++
	}
	t[key] = c
}

const noVerdict = "no verdict"

var (
	verdictOrder = []string{OutcomeKeep, OutcomeRewrite, OutcomeClose, noVerdict}
	reasonOrder  = []string{ReasonBuilt, ReasonFalsePremise, ReasonDoesNotApply, ReasonDuplicate, ReasonCouldNotVerify}
)

// summarize reads the rows of the night comments, oldest first. An
// acted verdict, one that is not a dry run, resets the "no verdict"
// count of its issue.
func summarize(nights []string) logSummary {
	s := logSummary{verdicts: tally{}, reasons: tally{}}
	failed := map[int]int{}
	for _, night := range nights {
		for line := range strings.Lines(night) {
			m := rowLine.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
			if m == nil {
				continue
			}
			issue, err := strconv.Atoi(m[1])
			if err != nil {
				continue // a number too long for an int names no issue
			}
			verdict, dryRun := m[2], m[3] == dryRunMark
			s.verdicts.add(verdict, dryRun)
			if verdict == OutcomeClose {
				// row writes a duplicate close as "duplicate of #N".
				reason, _, _ := strings.Cut(m[4], " of #")
				s.reasons.add(reason, dryRun)
			}
			switch {
			case verdict == noVerdict:
				failed[issue]++
			case !dryRun:
				delete(failed, issue)
			}
		}
	}
	for issue, n := range failed {
		if n >= noVerdictLimit {
			s.needsYou = append(s.needsYou, issue)
		}
	}
	slices.Sort(s.needsYou)
	return s
}

// markdown renders s for the tracking issue body.
func (s logSummary) markdown() string {
	var b strings.Builder
	b.WriteString("## Totals\n\n")
	b.WriteString("Rows in the night comments below. A dry-run row changed nothing on its issue.\n\n")
	b.WriteString("| Verdict | Rows | Dry run |\n|---|---|---|\n")
	for _, v := range verdictOrder {
		c := s.verdicts[v]
		dry := strconv.Itoa(c.dryRun)
		if v == noVerdict {
			dry = "-" // a no-verdict row acts on nothing, in any mode
		}
		fmt.Fprintf(&b, "| %s | %d | %s |\n", v, c.rows, dry)
	}
	b.WriteString("\n| Close reason | Rows | Dry run |\n|---|---|---|\n")
	for _, r := range reasonOrder {
		c := s.reasons[r]
		fmt.Fprintf(&b, "| %s | %d | %d |\n", r, c.rows, c.dryRun)
	}
	b.WriteString("\n## Needs you\n\n")
	fmt.Fprintf(&b, "The sweep skips these issues: each has %d \"no verdict\" rows since its last acted verdict.\n\n", noVerdictLimit)
	if len(s.needsYou) == 0 {
		b.WriteString("None.\n")
	}
	for _, issue := range s.needsYou {
		fmt.Fprintf(&b, "- #%d\n", issue)
	}
	return b.String()
}
