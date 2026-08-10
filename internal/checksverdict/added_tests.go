package checksverdict

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// testSuffix names a Go test file. It is the whole definition of "a
// test" here: the repair boundary has to be decidable from a path, and
// this is the only naming rule the toolchain itself enforces.
const testSuffix = "_test.go"

// AddedTestFiles returns the test files a pull request added, read from
// the name-status form of its diff — the output of
// `git diff --name-status` or `gh pr diff --name-status`, one record per
// line, status first.
//
// This decides the repair session's one hard boundary. A test the pull
// request added is the minion's own work, and a wrong assertion in it
// is the common failure this pipeline exists to fix. Every other test
// predates the pull request, and a session that may edit those can
// reach green by deleting what someone else proved.
//
// Only the "A" status counts. A modified test is a pre-existing test
// the pull request touched, and a renamed or copied one is old content
// at a new path — in both cases the assertions predate the pull
// request, so neither is an addition.
//
// The result is sorted and free of duplicates: a file added on one
// commit and re-added on another is still one file.
func AddedTestFiles(nameStatus string) []string {
	var added []string
	for line := range strings.SplitSeq(nameStatus, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "A" {
			continue
		}
		if path := fields[1]; strings.HasSuffix(path, testSuffix) {
			added = append(added, path)
		}
	}
	slices.Sort(added)
	return slices.Compact(added)
}

// WriteAddedTests saves the list where the repair session reads it, one
// path per line.
//
// An empty list is written as an empty file rather than skipped. The
// session's rule is "edit only the tests named here", so a missing file
// and an empty one must not look the same: the first is a crash
// upstream, and treating it as "no restriction" would hand the session
// every test in the repository.
func WriteAddedTests(path string, files []string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create the added-tests directory: %w", err)
		}
	}
	var b strings.Builder
	for _, f := range files {
		b.WriteString(f)
		b.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write the added-tests list: %w", err)
	}
	return nil
}
