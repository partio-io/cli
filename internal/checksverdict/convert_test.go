package checksverdict

import (
	"fmt"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/auditgate"
)

// Both commands succeeded, so the verdict is a pass with no findings
// and the gate takes it green without touching the pull request. A
// green check that still posts a comment would train everyone to
// ignore the comment.
func TestCleanRunPassesAndLeavesThePullRequestAlone(t *testing.T) {
	path := writeTo(t,
		Outcome{Name: "lint", Command: "make lint", Output: fixture(t, "lint-clean.txt")},
		Outcome{Name: "test", Command: "make test", Output: fixture(t, "test-clean.txt")},
	)
	gh := newFakeGitHub()

	res := runGate(t, gh, path)

	if !res.Green {
		t.Fatalf("gate returned red for a clean run: %+v", res)
	}
	if len(gh.created) != 0 || len(gh.updated) != 0 {
		t.Errorf("a clean run touched comments: created=%q updated=%v", gh.created, gh.updated)
	}
}

// The status comes from the exit code, so passing output that happens
// to contain failure-shaped text cannot turn the check red, and a
// command that failed silently cannot turn it green.
func TestStatusFollowsTheExitCodeNotTheOutput(t *testing.T) {
	cases := []struct {
		name      string
		outcome   Outcome
		want      string
		wantEmpty bool
	}{
		{
			name:      "failing output but the command succeeded",
			outcome:   Outcome{Name: "test", Command: "make test", Output: fixture(t, "test-fail.txt")},
			want:      "pass",
			wantEmpty: true,
		},
		{
			name:    "clean output but the command failed",
			outcome: Outcome{Name: "test", Command: "make test", Output: fixture(t, "test-clean.txt"), Failed: true},
			want:    "fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := Convert([]Outcome{tc.outcome})

			if v.Status != tc.want {
				t.Errorf("status = %q, want %q", v.Status, tc.want)
			}
			if tc.wantEmpty && len(v.Findings) != 0 {
				t.Errorf("a passing command produced %d findings: %+v", len(v.Findings), v.Findings)
			}
			if !tc.wantEmpty && len(v.Findings) == 0 {
				t.Error("a failing command produced no findings; the gate rejects that as malformed")
			}
		})
	}
}

// Every golangci-lint issue becomes its own finding, located at the
// exact file, line, and column the linter reported, and carrying both
// the message and the linter that raised it. The trailing summary
// block is bookkeeping, not an issue, and must not become a finding.
func TestLintIssuesBecomeOneFindingEach(t *testing.T) {
	v := Convert([]Outcome{
		{Name: "lint", Command: "make lint", Output: fixture(t, "lint-fail.txt"), Failed: true},
	})

	if v.Status != "fail" {
		t.Fatalf("status = %q, want fail", v.Status)
	}
	if len(v.Findings) != 4 {
		t.Fatalf("got %d findings, want 4 (one per lint issue): %+v", len(v.Findings), v.Findings)
	}
	for _, want := range []struct{ location, reasoning string }{
		{"internal/config/load.go:11:5", "Error return value of `os.Create` is not checked"},
		{"internal/config/load.go:12:15", "Error return value of `f.WriteString` is not checked"},
		{"internal/config/load.go:13:9", "Error return value of `f.Close` is not checked"},
		{"internal/config/load.go:8:6", "func unusedHelper is unused"},
	} {
		if !hasFinding(v.Findings, want.location, want.reasoning) {
			t.Errorf("no finding at %s saying %q:\n%+v", want.location, want.reasoning, v.Findings)
		}
	}
	// The linter's name is what tells a repair session which rule it
	// broke, so it has to survive into the reasoning.
	if !hasFinding(v.Findings, "load.go:8:6", "unused") {
		t.Error("findings do not name the linter that raised them")
	}
	for _, f := range v.Findings {
		if strings.Contains(f.Location, "issues:") || strings.HasPrefix(f.Location, "*") {
			t.Errorf("the summary block leaked into a finding: %+v", f)
		}
	}
}

// A package that never compiled prints compiler errors instead of test
// results, so there is no failing test name to hang them on. The
// finding still has to carry the file, the line, and the compiler's
// own words.
func TestBuildFailureCarriesTheCompilerErrors(t *testing.T) {
	v := Convert([]Outcome{
		{Name: "test", Command: "make test", Output: fixture(t, "test-build-fail.txt"), Failed: true},
	})

	if len(v.Findings) != 1 {
		t.Fatalf("got %d findings, want 1 (the package that failed to build): %+v", len(v.Findings), v.Findings)
	}
	f := v.Findings[0]
	if f.Location != "github.com/partio-io/cli/internal/attribution" {
		t.Errorf("location = %q, want the package that failed to build", f.Location)
	}
	for _, want := range []string{
		"internal/attribution/lines.go:3:37",
		"undefined: undefinedThing",
	} {
		if !strings.Contains(f.Reasoning, want) {
			t.Errorf("reasoning missing %q:\n%s", want, f.Reasoning)
		}
	}
	// The package that built and passed is not a finding.
	if strings.Contains(f.Location, "internal/config") {
		t.Errorf("a passing package produced a finding: %+v", f)
	}
}

// Subtests report twice: once inline as they run, and again as an
// indented roll-up under the parent. The roll-up is not detail, and a
// passing subtest's roll-up line must never be filed under a failing
// test — a reader who sees "--- PASS" under a failure stops trusting
// the report.
func TestPassingSubtestsDoNotPolluteAFailure(t *testing.T) {
	v := Convert([]Outcome{
		{Name: "test", Command: "make test", Output: fixture(t, "test-fail-subtests.txt"), Failed: true},
	})

	if len(v.Findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(v.Findings), v.Findings)
	}
	f := v.Findings[0]
	if !strings.Contains(f.Reasoning, `metadata session id = "", want "s-42"`) {
		t.Errorf("reasoning lost the assertion that failed:\n%s", f.Reasoning)
	}
	if strings.Contains(f.Reasoning, "PASS") {
		t.Errorf("a passing subtest was reported under the failure:\n%s", f.Reasoning)
	}
	if strings.Contains(f.Reasoning, "TestFindByBranch") {
		t.Errorf("a test that passed was named in the failure:\n%s", f.Reasoning)
	}
}

// When the subtests are what failed, the finding has to name them and
// carry their assertions. Naming only the parent sends a repair
// session to a table with no clue which row broke.
func TestFailingSubtestsAreNamedWithTheirAssertions(t *testing.T) {
	v := Convert([]Outcome{
		{Name: "test", Command: "make test", Output: fixture(t, "test-fail-subtest.txt"), Failed: true},
	})

	if len(v.Findings) != 1 {
		t.Fatalf("got %d findings, want 1 (one failing package): %+v", len(v.Findings), v.Findings)
	}
	f := v.Findings[0]
	for _, want := range []string{
		"TestLoadMergesLayers/local_over_repo",
		"TestLoadMergesLayers/env_over_local",
		`load_test.go:17: merged value = "global", want "local"`,
	} {
		if !strings.Contains(f.Reasoning, want) {
			t.Errorf("reasoning missing %q:\n%s", want, f.Reasoning)
		}
	}
	if strings.Contains(f.Reasoning, "repo_over_global") {
		t.Errorf("the subtest that passed was reported as failing:\n%s", f.Reasoning)
	}
	if strings.Contains(f.Reasoning, "this one is fine") {
		t.Errorf("a passing test's log output leaked into the failure:\n%s", f.Reasoning)
	}
}

// Output the parser does not recognise is the case that decides
// whether the gate can be trusted. It must never yield a pass, and
// never a fail with no findings — the gate rejects that as malformed,
// which would turn an unreadable failure into a fail-closed with the
// evidence thrown away. The raw output travels instead.
func TestUnparseableOutputStillFailsWithAnActionableFinding(t *testing.T) {
	v := Convert([]Outcome{
		{Name: "test", Command: "make test", Output: fixture(t, "unparseable.txt"), Failed: true},
	})

	if v.Status != "fail" {
		t.Fatalf("status = %q, want fail", v.Status)
	}
	if len(v.Findings) != 1 {
		t.Fatalf("got %d findings, want exactly 1 fallback: %+v", len(v.Findings), v.Findings)
	}
	f := v.Findings[0]
	if f.Location != "make test" {
		t.Errorf("location = %q, want the command that failed", f.Location)
	}
	if !strings.Contains(f.Reasoning, "No rule to make target") {
		t.Errorf("reasoning does not carry the raw output:\n%s", f.Reasoning)
	}
}

// Whatever a command prints, a failure must arrive as findings the
// gate accepts. Anything else fails the check closed and discards the
// reason it failed.
func TestAFailedCommandAlwaysProducesUsableFindings(t *testing.T) {
	outputs := []string{
		fixture(t, "test-fail.txt"),
		fixture(t, "test-build-fail.txt"),
		fixture(t, "lint-fail.txt"),
		fixture(t, "unparseable.txt"),
		"",
		"FAIL\n",
		"panic: runtime error: index out of range [3] with length 2",
	}
	for _, name := range []string{"lint", "test", "something-else"} {
		for i, out := range outputs {
			t.Run(fmt.Sprintf("%s/%d", name, i), func(t *testing.T) {
				v := Convert([]Outcome{{Name: name, Command: "make " + name, Output: out, Failed: true}})

				if v.Status != "fail" {
					t.Fatalf("status = %q, want fail", v.Status)
				}
				if len(v.Findings) == 0 {
					t.Fatal("a failed command produced no findings; the gate would fail closed and lose the reason")
				}
				for _, f := range v.Findings {
					if strings.TrimSpace(f.Location) == "" || strings.TrimSpace(f.Reasoning) == "" {
						t.Errorf("finding is missing location or reasoning: %+v", f)
					}
				}
			})
		}
	}
}

// hasFinding reports whether some finding matches both substrings.
func hasFinding(findings []auditgate.Finding, location, reasoning string) bool {
	for _, f := range findings {
		if strings.Contains(f.Location, location) && strings.Contains(f.Reasoning, reasoning) {
			return true
		}
	}
	return false
}
