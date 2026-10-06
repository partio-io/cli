package auditgate

import (
	"bytes"
	"fmt"

	"github.com/partio-io/cli/internal/github"
)

// commentPrefix is the shared body prefix for all minion audit
// comments. The gate appends the audit name so different audits on
// the same PR each keep their own comment.
const commentPrefix = "Minion audit — "

// failBody renders the single red-run comment from a fail verdict.
func failBody(auditName string, v Verdict) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s%s audit failed.\n", commentPrefix, auditName)
	writeFindings(&b, v)
	return b.String()
}

// writeFindings renders the numbered findings list both red-run
// comments share.
func writeFindings(b *bytes.Buffer, v Verdict) {
	for i, f := range v.Findings {
		fmt.Fprintf(b, "\n%d. `%s` — %s\n", i+1, f.Location, f.Reasoning)
	}
}

// plural picks the form matching n. The comment is read by a person
// deciding whether to take the pull request over, so it reads as
// prose rather than as log output.
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// exhaustedBody renders the comment for a fail verdict that no further
// repair round will address, because rounds is the whole budget. It
// opens differently from failBody on purpose: an operator reading the
// pull request must be able to tell "gave up after spending the
// budget" from "failed on the first pass" without opening a run log.
func exhaustedBody(auditName string, v Verdict, rounds int) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s%s audit gave up after %d repair %s; the repair budget is spent.\n",
		commentPrefix, auditName, rounds, plural(rounds, "round", "rounds"))
	if len(v.Findings) == 1 {
		fmt.Fprint(&b, "\nThis finding survived every round:\n")
	} else {
		fmt.Fprintf(&b, "\nThese %d findings survived every round:\n", len(v.Findings))
	}
	writeFindings(&b, v)
	fmt.Fprint(&b, "\nNo further repair round starts for this check on this branch. "+
		"The pull request stays red and open for a human to take over.\n")
	return b.String()
}

// failClosedBody renders the red-run comment for a run whose verdict
// was missing or malformed: the audit produced nothing the gate can
// trust, so the gate reports that instead of findings.
func failClosedBody(auditName string, cause error) string {
	return fmt.Sprintf("%s%s audit produced no usable verdict; failing closed.\n\nDetail: %v\n",
		commentPrefix, auditName, cause)
}

// upsertComment creates the audit comment on the PR, or updates the
// existing one in place, matched by body prefix.
func upsertComment(cfg Config, body string) error {
	return cfg.client().UpsertComment(cfg.Repo, cfg.PRNumber, commentPrefix+cfg.AuditName, body)
}

// client is the GitHub client cfg describes.
func (c Config) client() github.Client {
	return github.Client{BaseURL: c.APIBaseURL, Token: c.Token, HTTPClient: c.HTTPClient}
}
