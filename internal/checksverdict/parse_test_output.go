package checksverdict

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/partio-io/cli/internal/auditgate"
)

// The verbose test format, matched structurally rather than by raw
// prefix: an assertion message can itself start with "---" (a unified
// diff does), and that message is the most valuable line in the
// report.
var (
	testRun    = regexp.MustCompile(`^=== RUN\s+(\S+)$`)
	testStatus = regexp.MustCompile(`^--- (PASS|FAIL|SKIP|BENCH): (\S+)`)
	testOther  = regexp.MustCompile(`^=== (PAUSE|CONT|NAME)\s`)
)

const (
	// A failing package's report ends up in a pull request comment, so
	// each part of it is bounded: the lines kept under one test, the
	// compiler errors or panic dump kept under a package, and the
	// number of failing tests named before the rest are counted.
	maxDetailLines = 20
	maxLooseLines  = 40
	maxFailedTests = 10
)

// parseTestOutput turns `go test ./... -v` output into one finding per
// failing package.
//
// The package a line belongs to is only known at the line that closes
// it — "ok" or "FAIL" — so lines accumulate until one of those arrives
// and are then either dropped or attributed.
func parseTestOutput(out, label string) []auditgate.Finding {
	var (
		findings []auditgate.Finding
		buf      []string
	)
	for _, line := range strings.Split(out, "\n") {
		switch {
		case closesPassingPackage(line):
			buf = nil
		case strings.HasPrefix(line, "FAIL\t"):
			if pkg := packageOf(line); pkg != "" {
				findings = append(findings, auditgate.Finding{
					Location:  pkg,
					Reasoning: failureReport(pkg, label, buf),
				})
			}
			buf = nil
		case line == "PASS" || line == "FAIL":
			// Per-package separators; they carry nothing.
		default:
			buf = append(buf, line)
		}
	}
	return findings
}

// closesPassingPackage reports whether the line ends a package that
// needs no finding: one that passed, or one with no test files.
func closesPassingPackage(line string) bool {
	return strings.HasPrefix(line, "ok  \t") || strings.HasPrefix(line, "?   \t")
}

// packageOf pulls the import path out of a "FAIL" line. It covers both
// "FAIL\tpkg\t0.01s" and "FAIL\tpkg [build failed]".
func packageOf(line string) string {
	fields := strings.Fields(strings.TrimPrefix(line, "FAIL"))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// failureReport builds the reasoning for one failing package: which
// tests failed and what they printed, or the compiler errors when the
// package never built. It is written to be read without re-running the
// command.
func failureReport(pkg, label string, buf []string) string {
	failed, loose := splitFailures(buf)
	report := &strings.Builder{}
	fmt.Fprintf(report, "Package %s failed under `%s`.", pkg, label)
	shown := failed
	if len(shown) > maxFailedTests {
		shown = shown[:maxFailedTests]
	}
	for _, f := range shown {
		fmt.Fprintf(report, "\n\n%s failed:", f.test)
		for _, d := range f.detail {
			fmt.Fprintf(report, "\n%s", strings.TrimRight(d, " \t"))
		}
	}
	if rest := len(failed) - len(shown); rest > 0 {
		fmt.Fprintf(report, "\n\n… and %d more failing test(s) in this package.", rest)
	}
	if len(loose) > 0 {
		// A build failure prints compiler errors instead of test
		// results, and a panic prints a stack; neither has a test name
		// to hang it on.
		fmt.Fprintf(report, "\n\n%s", strings.Join(loose, "\n"))
	}
	return report.String()
}

// testFailure is one failing test and the lines it printed.
type testFailure struct {
	test   string
	detail []string
}

// splitFailures separates a failing package's buffered lines into the
// tests that failed with their output, and everything that belonged to
// no test — compiler errors, panics, and stray writes to stderr.
//
// Output is attributed by test name, not by position: subtests report
// once inline as they run and again as an indented roll-up under the
// parent, so the run that produced a line and the line that declares
// it failed can be far apart.
func splitFailures(buf []string) (failed []testFailure, loose []string) {
	var (
		detail  = map[string][]string{}
		order   []string
		current string
	)
	for _, line := range buf {
		trimmed := strings.TrimSpace(line)
		switch {
		case testRun.MatchString(trimmed):
			current = testRun.FindStringSubmatch(trimmed)[1]
		case testStatus.MatchString(trimmed):
			m := testStatus.FindStringSubmatch(trimmed)
			if m[1] == "FAIL" {
				order = append(order, m[2])
			}
			current = ""
		case testOther.MatchString(trimmed):
			// PAUSE, CONT and NAME are scheduling noise.
		case trimmed == "":
		case current != "":
			if len(detail[current]) < maxDetailLines {
				detail[current] = append(detail[current], line)
			}
		default:
			if len(loose) < maxLooseLines {
				loose = append(loose, line)
			}
		}
	}
	for _, name := range withoutRolledUpParents(order) {
		failed = append(failed, testFailure{test: name, detail: detail[name]})
	}
	return failed, loose
}

// withoutRolledUpParents drops a parent test whose failure is only the
// sum of its subtests' failures. The subtest names say which table row
// broke; the parent's name says nothing the children do not.
func withoutRolledUpParents(names []string) []string {
	kept := make([]string, 0, len(names))
	for _, name := range names {
		if hasFailingChild(name, names) {
			continue
		}
		kept = append(kept, name)
	}
	return kept
}

func hasFailingChild(name string, names []string) bool {
	for _, other := range names {
		if other != name && strings.HasPrefix(other, name+"/") {
			return true
		}
	}
	return false
}
