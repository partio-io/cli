// Command minion-checks-verdict runs the repository's own lint and
// test targets and writes their result as a verdict file, in the shape
// the minion audits emit.
//
// It does not decide the check's outcome. minion-audit-gate reads the
// verdict and owns red or green, exactly as it does for an audit, so
// this command exits 0 whenever it managed to write a verdict — even
// one that says the checks failed. A failure to write is the only
// error, and the gate then fails the check closed on its own.
//
// With --pr-files it does a different, smaller job: it reads a
// name-status diff and writes the list of test files the pull request
// added. That list is the repair session's boundary — the only tests it
// may edit — so it is decided here and handed over, not left to the
// session to work out.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/checksverdict"
)

func main() {
	var (
		dir     = flag.String("dir", ".", "directory to run the checks in")
		out     = flag.String("out", ".minion-checks/verdict.json", "path to write the verdict to")
		prFiles = flag.String("pr-files", "",
			"name-status diff of the pull request; writes the added-test list instead of running the checks")
		testsOut = flag.String("tests-out", ".minion-checks/added-tests.txt",
			"path to write the added-test list to, with --pr-files")
	)
	flag.Parse()

	if *prFiles != "" {
		if err := writeAddedTests(*prFiles, *testsOut); err != nil {
			fmt.Fprintf(os.Stderr, "minion-checks-verdict: %v\n", err)
			os.Exit(1)
		}
		return
	}

	verdict := checksverdict.Convert(checksverdict.Run(*dir, checksverdict.DefaultCommands()))
	if err := checksverdict.Write(*out, verdict); err != nil {
		fmt.Fprintf(os.Stderr, "minion-checks-verdict: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("minion-checks-verdict: %s with %d finding(s) → %s\n",
		verdict.Status, len(verdict.Findings), *out)
}

// writeAddedTests reads a name-status diff and saves the test files the
// pull request added.
//
// An unreadable diff is an error, not an empty list. The list bounds
// what the repair session may edit, and a silent empty one would read
// as "this pull request added no tests" when the truth is "nobody
// looked".
func writeAddedTests(prFiles, testsOut string) error {
	diff, err := os.ReadFile(filepath.Clean(prFiles))
	if err != nil {
		return fmt.Errorf("read the pull request file list: %w", err)
	}
	added := checksverdict.AddedTestFiles(string(diff))
	if err := checksverdict.WriteAddedTests(testsOut, added); err != nil {
		return err
	}
	fmt.Printf("minion-checks-verdict: the pull request added %d test file(s) → %s\n", len(added), testsOut)
	return nil
}
