package checksverdict

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/partio-io/cli/internal/auditgate"
)

// lintIssue matches the head of one golangci-lint issue:
// "path/file.go:12:34: message (linter)". The column is optional. The
// path must carry a file extension, which is what keeps the trailing
// summary block and any log lines from parsing as issues.
var lintIssue = regexp.MustCompile(`^(\S+\.\w+:\d+(?::\d+)?): (.+)$`)

// lintSummary matches the count line that closes the report, after
// which nothing belongs to an issue any more.
var lintSummary = regexp.MustCompile(`^\d+ issues?:$`)

// maxLintContext caps the source lines kept under one issue. The
// reporter prints the offending line and a caret; more than a few
// lines means the format changed, and a pull request comment is not
// the place to find that out.
const maxLintContext = 4

// parseLintOutput turns golangci-lint output into one finding per
// issue. The source excerpt the reporter prints under each issue is
// kept, so a repair session can see the offending line without opening
// the file.
func parseLintOutput(out, label string) []auditgate.Finding {
	var (
		findings []auditgate.Finding
		context  []string
	)
	flush := func() {
		if len(findings) == 0 || len(context) == 0 {
			context = nil
			return
		}
		last := &findings[len(findings)-1]
		last.Reasoning += "\n" + strings.Join(context, "\n")
		context = nil
	}
	for _, line := range strings.Split(out, "\n") {
		switch {
		case lintSummary.MatchString(strings.TrimSpace(line)):
			flush()
			return findings
		case lintIssue.MatchString(line):
			flush()
			m := lintIssue.FindStringSubmatch(line)
			findings = append(findings, auditgate.Finding{
				Location:  m[1],
				Reasoning: fmt.Sprintf("`%s` reported: %s", label, m[2]),
			})
		case strings.TrimSpace(line) == "":
		case len(context) < maxLintContext:
			context = append(context, strings.TrimRight(line, " \t"))
		}
	}
	flush()
	return findings
}
