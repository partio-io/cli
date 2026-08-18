package programshape

import (
	"path/filepath"
	"slices"
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

	// The three programs below are one-shot feature specifications, not
	// pipeline programs. The PRD for this work did not count them. Each has a
	// bare H1 with no prose under it, so the model receives a title and
	// nothing else — a worse state than the five programs above.
	"e2e-model-name-checkpoint-flow.md":    "The specification sits in ## Background and ## What to build; the H1 carries no prose, so the model receives only the title.",
	"e2e-squash-merge-branch-name-flow.md": "The specification sits in ## What to build; the H1 carries no prose, so the model receives only the title and ## Context.",
	"e2e-two-rapid-commit-trailer-flow.md": "The specification sits in ## Background and ## What to build; the H1 carries no prose, so the model receives only the title.",
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
