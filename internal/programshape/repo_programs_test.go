package programshape

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// programsDir is this repository's minions programs directory, relative to the
// package directory that go test runs in.
const programsDir = "../../.minions/programs"

// knownUnreachable records the programs whose instructions do not reach the
// model today. Every entry is a defect that is accepted, not approved: the
// suite stays green on arrival and the list is the checklist to work down.
// Delete an entry when its program is fixed — the test fails if a listed
// program turns out to be healthy, so the list cannot drift from reality.
var knownUnreachable = map[string]string{
	"approve.md":       "The whole procedure sits in ## Steps; the program is unreferenced by any workflow and is out of scope for this work.",
	"ingest.md":        "The whole procedure sits in ## Steps; the program is unreferenced by any workflow and is out of scope for this work.",
	"doc-update.md":    "The whole procedure sits in ## Steps; the program is reachable from a workflow, so it misbehaves in production, and it is out of scope for this work.",
	"readme-update.md": "The whole procedure sits in ## Steps; the program is unreferenced by any workflow and is out of scope for this work.",
}

func TestRepoProgramsKeepInstructionsReachable(t *testing.T) {
	scan, err := ScanDir(programsDir)
	if err != nil {
		t.Fatalf("scan the repository programs: %v", err)
	}
	if len(scan.Programs) == 0 {
		t.Fatalf("no .md programs found in %s", programsDir)
	}

	for name, findings := range scan.Unreachable {
		if _, accepted := knownUnreachable[name]; !accepted {
			t.Errorf("%s", Report(filepath.Join(programsDir, name), findings))
		}
	}

	for name := range knownUnreachable {
		switch {
		case !slices.Contains(scan.Programs, name):
			t.Errorf("knownUnreachable lists %s, but no such program exists in %s — "+
				"delete the stale entry.", name, programsDir)
		case len(scan.Unreachable[name]) == 0:
			t.Errorf("%s is listed in knownUnreachable, but its instructions now reach "+
				"the model — delete its entry from knownUnreachable.", name)
		}
	}
}

// runtimeHeading is the heading pattern of the minions program parser
// (internal/program/parse.go, headingRe, at v0.0.14). The parser applies it to
// every line and knows nothing of fences, so a heading line inside a fenced
// example still opens a section there, and the agent loses the rest of its
// instructions. Check skips fenced lines, so it cannot see that cut. An
// indented line does not match, which is why a fenced example inside a list
// item is safe.
var runtimeHeading = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// TestRepoProgramsShowNoHeadingInAFence guards the fenced examples in the
// programs: the review program shows a whole issue body, with its ## headings,
// inside the reviewer's instructions.
func TestRepoProgramsShowNoHeadingInAFence(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(programsDir, "*.md"))
	if err != nil {
		t.Fatalf("list the repository programs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no .md programs found in %s", programsDir)
	}

	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		inFence := false
		for n, line := range strings.Split(string(raw), "\n") {
			if isFence(line) {
				inFence = !inFence
				continue
			}
			if inFence && runtimeHeading.MatchString(line) {
				t.Errorf("%s:%d: %q sits in a fence, but the runtime reads it as a heading and ends the section there; indent the fenced block", path, n+1, line)
			}
		}
	}
}
