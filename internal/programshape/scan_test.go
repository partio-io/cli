package programshape

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writePrograms fills an isolated programs directory and returns its path.
func writePrograms(t *testing.T, programs map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, src := range programs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatalf("write program %s: %v", name, err)
		}
	}
	return dir
}

const healthyProgram = `# Repair the failing check

## Agents

### checks-fix

1. Read the failing check output.
2. Repair the failure and push the fix.
`

const stepsProgram = `# Release the new version

Cut a release.

## Steps

1. Tag the commit.
2. Push the tag.
`

// A new program that hides its procedure in ## Steps must turn the suite red,
// which is the whole point of the check.
func TestScanDirFlagsANewStepsProgram(t *testing.T) {
	dir := writePrograms(t, map[string]string{
		"checks-fix.md": healthyProgram,
		"release.md":    stepsProgram,
	})

	scan, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	if _, ok := scan.Unreachable["checks-fix.md"]; ok {
		t.Errorf("checks-fix.md was flagged, but its instructions reach the model")
	}

	findings, ok := scan.Unreachable["release.md"]
	if !ok {
		t.Fatalf("release.md was not flagged, but its procedure sits in a dropped ## Steps section")
	}
	if len(findings) != 1 || findings[0].Heading != "## Steps" {
		t.Errorf("release.md findings = %v, want one finding for %q", findings, "## Steps")
	}
}

func TestScanDirReadsEveryMarkdownProgramAndNothingElse(t *testing.T) {
	dir := writePrograms(t, map[string]string{
		"release.md":    stepsProgram,
		"checks-fix.md": healthyProgram,
		"notes.txt":     stepsProgram,
		"sources.yaml":  "id: release\n",
	})

	scan, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	want := []string{"checks-fix.md", "release.md"}
	if !slices.Equal(scan.Programs, want) {
		t.Errorf("ScanDir read %v, want %v", scan.Programs, want)
	}
}

func TestScanDirReportsAMissingDirectory(t *testing.T) {
	if _, err := ScanDir(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("ScanDir returned no error for a missing directory")
	}
}

// A failure must name the program file and the dropped heading, so the fix
// needs no bisecting of the program format rules.
func TestReportNamesTheProgramAndTheDroppedHeading(t *testing.T) {
	path := filepath.Join(".minions", "programs", "release.md")
	message := Report(path, Check(stepsProgram))

	if !strings.Contains(message, path) {
		t.Errorf("message does not name the program file %q:\n%s", path, message)
	}
	if !strings.Contains(message, "## Steps") {
		t.Errorf("message does not name the dropped heading %q:\n%s", "## Steps", message)
	}
}

func TestReportNamesEveryDroppedHeading(t *testing.T) {
	src := `# Build an end-to-end test

## Background

1. The trailer is written by the post-commit hook.

## What to build

1. Create a repository fixture.
`
	message := Report("e2e.md", Check(src))

	for _, heading := range []string{"## Background", "## What to build"} {
		if !strings.Contains(message, heading) {
			t.Errorf("message does not name the dropped heading %q:\n%s", heading, message)
		}
	}
}
