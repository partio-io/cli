package review

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/partio-io/cli/internal/premise"
)

// criteriaHeading starts the acceptance criteria of an issue body.
const criteriaHeading = "## Acceptance Criteria"

// proposalIDLine is the line that ends a proposal body and names its id.
var proposalIDLine = regexp.MustCompile(`^Proposal id: \S+$`)

// programPointer is the link an older proposal carried to its proposal
// file. Proposal files no longer exist, so the pointer leads nowhere.
var programPointer = regexp.MustCompile(`<!--\s*program:`)

// maxIssueBody is GitHub's limit on an issue body. GitHub counts
// characters; the gate counts bytes, which is never fewer.
const maxIssueBody = 65536

// checkShape rejects a rewrite whose new text does not have the issue
// shape the proposer files today, so a body that fails it never
// replaces a working one. oldBody is the body the rewrite replaces.
func checkShape(oldBody string, rw Rewrite) error {
	if len(rw.Body) > maxIssueBody {
		return fmt.Errorf("rewrite body has %d bytes, past GitHub's limit of %d on an issue body", len(rw.Body), maxIssueBody)
	}
	if _, err := premise.Parse(rw.Body); err != nil {
		return fmt.Errorf("rewrite body has no premise block the premise package accepts: %w", err)
	}
	if !hasCriterion(rw.Body) {
		return fmt.Errorf("rewrite body has no %q section with a \"- [ ]\" item", criteriaHeading)
	}
	if !proposalIDLine.MatchString(lastLine(rw.Body)) {
		return errors.New(`rewrite body does not end with a "Proposal id: <id>" line`)
	}
	if programPointer.MatchString(rw.Body) {
		return errors.New("rewrite body still carries a <!-- program: … --> pointer to a proposal file")
	}
	kept := SourceRefs(rw.Body)
	for _, ref := range SourceRefs(oldBody) {
		if !slices.Contains(kept, ref) {
			return fmt.Errorf("rewrite body drops the source reference %s of the old body", ref)
		}
	}
	return nil
}

// hasCriterion reports whether body has an acceptance criteria section,
// from its heading to the next heading of the same level, with at least
// one unchecked checklist item.
func hasCriterion(body string) bool {
	in := false
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.EqualFold(line, criteriaHeading):
			in = true
		case strings.HasPrefix(line, "## "):
			in = false
		case in && strings.HasPrefix(line, "- [ ] "):
			return true
		}
	}
	return false
}

// lastLine returns the last line of text that is not blank, trimmed.
func lastLine(text string) string {
	text = strings.TrimSpace(text)
	return strings.TrimSpace(text[strings.LastIndex(text, "\n")+1:])
}
