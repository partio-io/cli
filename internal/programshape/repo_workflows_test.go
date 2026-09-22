package programshape

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// workflowsDir is this repository's GitHub workflows directory, relative to
// the package directory that go test runs in.
const workflowsDir = "../../.github/workflows"

// programRef matches a program path in a workflow file.
var programRef = regexp.MustCompile(`\.minions/programs/([A-Za-z0-9._-]+\.md)`)

// manualPrograms records the programs that no workflow runs. Every other file
// in the programs directory must be one a workflow runs. A proposal is never a
// file here: it is a GitHub issue, because a build reads the issue and nothing
// reads a proposal file.
var manualPrograms = map[string]string{
	"approve.md":       "No workflow runs it.",
	"ingest.md":        "No workflow runs it.",
	"readme-update.md": "No workflow runs it.",
}

// TestRepoProgramsAreRunByAWorkflow keeps the programs directory to programs.
// On 2026-09-22 it held 11 proposal specifications beside the 9 programs the
// workflows run. No build read them, and the shape check spent its failures on
// files that no model ever receives.
func TestRepoProgramsAreRunByAWorkflow(t *testing.T) {
	scan, err := ScanDir(programsDir)
	if err != nil {
		t.Fatalf("scan the repository programs: %v", err)
	}
	run := runByAWorkflow(t)

	for _, name := range scan.Programs {
		_, manual := manualPrograms[name]
		switch {
		case run[name] && manual:
			t.Errorf("manualPrograms lists %s, but a workflow runs it — delete its entry.", name)
		case !run[name] && !manual:
			t.Errorf("%s is in %s, but no workflow runs it. A proposal belongs in its GitHub "+
				"issue, which is all a build reads. If it is a program, run it from a workflow "+
				"or list it in manualPrograms with a reason.", name, programsDir)
		}
	}

	for name := range manualPrograms {
		if !slices.Contains(scan.Programs, name) {
			t.Errorf("manualPrograms lists %s, but no such program exists in %s — "+
				"delete the stale entry.", name, programsDir)
		}
	}
}

// TestRepoProgramsStageNoProposalFile pins how proposal files arrived: a
// program a workflow runs told its agent to commit into the programs
// directory. No such program may stage that directory.
func TestRepoProgramsStageNoProposalFile(t *testing.T) {
	scan, err := ScanDir(programsDir)
	if err != nil {
		t.Fatalf("scan the repository programs: %v", err)
	}

	for name := range runByAWorkflow(t) {
		if !slices.Contains(scan.Programs, name) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(programsDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for line := range strings.SplitSeq(string(src), "\n") {
			if strings.Contains(line, "git add") && strings.Contains(line, ".minions/programs") {
				t.Errorf("%s stages the programs directory, so its run commits a proposal file:\n%s",
					name, strings.TrimSpace(line))
			}
		}
	}
}

// runByAWorkflow reports the program file names that the repository's
// workflows run.
func runByAWorkflow(t *testing.T) map[string]bool {
	t.Helper()

	var paths []string
	for _, pattern := range []string{"*.yml", "*.yaml"} {
		matches, err := filepath.Glob(filepath.Join(workflowsDir, pattern))
		if err != nil {
			t.Fatalf("list the workflows: %v", err)
		}
		paths = append(paths, matches...)
	}

	run := map[string]bool{}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, m := range programRef.FindAllStringSubmatch(string(src), -1) {
			run[m[1]] = true
		}
	}
	if len(run) == 0 {
		t.Fatalf("no workflow in %s runs a program", workflowsDir)
	}
	return run
}
