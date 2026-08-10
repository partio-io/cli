// Package checksverdict turns the repository's own lint and test
// commands into the verdict shape the minion audits already emit. The
// gate, the pull request comment, the round accounting, and the patch
// applier then work on a failing test exactly as they work on a
// dead-code finding, without knowing which produced it.
package checksverdict

import (
	"fmt"
	"strings"

	"github.com/partio-io/cli/internal/auditgate"
)

// maxRawOutput caps the raw output carried by a fallback finding. The
// tail is the useful end of a command's output, and a pull request
// comment has to stay readable.
const maxRawOutput = 2000

// Outcome is one check command's captured result. Name both labels the
// check in its findings and selects the parser: "lint" and "test" are
// understood, and anything else falls back to the raw output.
type Outcome struct {
	Name    string
	Command string // what ran, e.g. "make test", quoted in fallback findings
	Output  string // combined stdout and stderr
	Failed  bool   // the command exited non-zero
}

// Convert maps captured outcomes to a verdict. The command's exit
// status decides the status — never the parse, which can only add
// detail to a failure the command already reported. A failed command
// whose output yields nothing still produces a finding, because a fail
// verdict with no findings is one the gate rejects as malformed.
func Convert(outcomes []Outcome) auditgate.Verdict {
	v := auditgate.Verdict{Status: "pass", Findings: []auditgate.Finding{}}
	for _, o := range outcomes {
		if !o.Failed {
			continue
		}
		v.Status = "fail"
		found := o.findings()
		if len(found) == 0 {
			found = []auditgate.Finding{o.unreadable()}
		}
		v.Findings = append(v.Findings, found...)
	}
	return v
}

// findings parses one failed command's output.
func (o Outcome) findings() []auditgate.Finding {
	switch o.Name {
	case "test":
		return parseTestOutput(o.Output, o.label())
	case "lint":
		return parseLintOutput(o.Output, o.label())
	default:
		return nil
	}
}

// unreadable is the finding for a command that failed in a way its
// parser does not recognise — a missing make target, a runner that
// died, a format change. It hands over the raw tail so the failure is
// still actionable, and it names the command so it can be re-run.
func (o Outcome) unreadable() auditgate.Finding {
	return auditgate.Finding{
		Location: o.label(),
		Reasoning: fmt.Sprintf(
			"%s failed, and its output does not parse into findings. Raw output:\n%s",
			o.label(), tail(o.Output)),
	}
}

// label names the command in a finding, preferring what actually ran.
func (o Outcome) label() string {
	if o.Command != "" {
		return o.Command
	}
	if o.Name != "" {
		return o.Name
	}
	return "check"
}

// tail returns the last maxRawOutput characters, marking the cut.
func tail(out string) string {
	out = strings.TrimSpace(out)
	if out == "" {
		return "(no output)"
	}
	if len(out) <= maxRawOutput {
		return out
	}
	return "…\n" + out[len(out)-maxRawOutput:]
}
