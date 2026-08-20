# First-Commit Attribution

## Problem Statement

**Partio records zero attributed lines for the first commit in a
repository.** A user who installs Partio and commits gets a checkpoint
that claims the commit added nothing. The earliest work in a repository
is the work most likely to be large, and it is the work Partio counts
as empty.

The cause is a corrupted constant. The attribution code asks git for the
line counts of a commit. It compares the commit against its parent. A
first commit has no parent, so that comparison fails. The code then
falls back to a comparison against git's empty tree object, and the
constant it uses for that object is
`4b825dc642cb6eb9a060e54bf899d69f82cf7ee2`. The real empty tree object
is `4b825dc642cb6eb9a060e54bf8d69288fbee4904`. Git answers the fallback
with `fatal: bad object`. The attribution code catches that error,
returns an empty result, and reports no error to its caller. The post-
commit hook therefore has nothing to log and nothing to warn about.

**The same defect makes a first checkpoint hold no diff.** Three
functions in the git package identify a commit's content by a comparison
against the parent commit. All three fail on a first commit. One
produces the line counts, one produces the changed file list, and one
produces the unified diff that the checkpoint stores. A user's first
checkpoint therefore holds zero lines, no file list, and no diff.

**A speed claim about this code path cannot be checked.** Issue #30 said
the post-commit hook performed a full tree walk, and that a replacement
would cut hook latency. The hook never performed a tree walk. The claim
came from a sibling product. The repository holds no benchmark, so the
claim went unmeasured for four months and a pull request was built on
it.

## Solution

Partio counts the lines in a first commit, lists its files, and stores
its diff, exactly as it does for every later commit.

One place in the git package decides how to identify a commit's content.
A commit with a parent keeps today's behaviour, without any change. A
commit with no parent is compared against git's empty tree object, under
a named constant that no caller can mistype.

Attribution stops the report of a failed measurement as a zero
measurement. When git cannot answer, the caller learns about it and
writes a warning. Hooks stay non-blocking, as the repository requires.

A benchmark compares the current comparison strategy against the
alternative that issue #30 proposed. The next person who claims that a
change to this path is faster must show a number. The benchmark also
records the outcome for this specific alternative: it is not faster.

## User Stories

1. As a Partio user, I want my first commit to record its attributed lines, so that my earliest work is not silently counted as zero.
2. As a Partio user, I want my first checkpoint to store the changed file list, so that the checkpoint describes what the commit touched.
3. As a Partio user, I want my first checkpoint to store the unified diff, so that the checkpoint holds the code and not only the session.
4. As a Partio user, I want a fresh repository to behave like an established one, so that I do not need a second commit before Partio works.
5. As a Partio user, I want a repository with exactly one commit to report a correct total, so that a new project is not misreported from day one.
6. As a Partio user, I want a first commit that adds many files to count all of them, so that a project import is measured.
7. As a Partio user, I want a first commit that adds a binary file to skip that file and still count the text files, so that one binary does not zero the commit.
8. As a Partio user, I want a first empty commit to record zero lines truthfully, so that a real zero and a failed measurement are different outcomes.
9. As a Partio user, I want merge commits to keep the counts they have today, so that a fix for first commits does not change my history.
10. As a Partio user, I want a merge commit to report the lines it brings in against its first parent, so that merge attribution stays comparable across releases.
11. As a Partio user, I want a regular commit to keep the counts it has today, so that the common path carries no risk from this change.
12. As a Partio user, I want an agent-authored first commit to record one hundred percent agent lines, so that agent attribution starts at the first commit.
13. As a Partio user, I want a human-authored first commit to record zero percent agent lines, so that my own work is attributed to me.
14. As a Partio user, I want a failed measurement to appear in the log, so that I can tell a broken measurement from an empty commit.
15. As a Partio user, I want a failed measurement to leave my commit intact, so that a hook problem never blocks my work.
16. As a Partio user, I want the empty tree object identified in one named place, so that a typed literal cannot break this path again.
17. As a Partio user, I want the fix to cover every commit-diff function, so that one function is not left broken after the others are fixed.
18. As a Partio user, I want my post-commit path to keep its current speed, so that a correctness fix does not cost me latency.
19. As a Partio user, I want hook changes justified by measurement, so that my post-commit path does not get slower in the name of speed.
20. As the operator, I want a benchmark that compares against the old behaviour, so that "faster" is demonstrated rather than asserted.
21. As the operator, I want the benchmark to cover a repository with many files, so that a scale claim is tested at scale.
22. As the operator, I want the benchmark to hold both strategies side by side, so that a comparison needs no second branch.
23. As the operator, I want the rejected strategy kept out of production code, so that the shipped code carries exactly one strategy.
24. As the operator, I want the benchmark result written down, so that the next proposal about this path meets a recorded number.
25. As the operator, I want the false premise of issue #30 recorded here, so that the same claim does not return unchallenged.
26. As the operator, I want the merge regression in pull request #653 recorded here, so that the reason to reject that pull request is written and not remembered.
27. As the operator, I want the premise-gate document corrected, so that our own record of this defect states the real cause.
28. As the operator, I want pull request #653 left open, so that the decision to close it stays mine.
29. As the operator, I want every changed function covered by a test, so that a later refactor cannot reintroduce this defect quietly.
30. As the operator, I want the tests to follow the house pattern, so that a reader of this package finds nothing unfamiliar.
31. As a Partio contributor, I want the root-commit decision made in one function, so that a future diff helper inherits the fix.
32. As a Partio contributor, I want the tests to build real repositories, so that the tests measure git and not a mock of git.
33. As a Partio contributor, I want a first-commit case in the test table of each function, so that the defect has a permanent guard.
34. As a Partio contributor, I want a merge-commit case in the test table of each function, so that the regression in pull request #653 cannot land later.

## Implementation Decisions

**Language.** Go. The repository is Go, and an existing repository's
language wins. No new service and no new script is added.

**One decision point for the commit range.** The git package gets an
unexported helper. It takes a commit hash. It returns the two revision
arguments that identify that commit's content:

- For a commit with a parent, it returns the parent revision and the
  commit revision. This is today's behaviour, unchanged.
- For a commit with no parent, it returns the empty tree constant and
  the commit revision.

The helper detects the parent through git, not through a guess. It does
not swallow an error from git. An error that is not the absence of a
parent propagates to the caller.

**A named empty tree constant.** The git package holds one exported
constant for git's empty tree object id, with a comment that states what
it is. The corrupted literal is deleted. No other package holds a copy.

**Three public functions keep their signatures.** The function for line
counts, the function for the changed file list, and the function for the
unified diff each build their own flags and then append the two
revisions from the helper. Their signatures do not change, so no caller
changes.

**Merge semantics do not change.** A merge commit has a parent, so it
takes the unchanged path and is compared against its first parent. The
`--root` flag is not added to production code, because it makes a merge
commit produce no output at all.

**Attribution stops the silent zero.** The attribution function drops
its private copy of the empty tree fallback. When git cannot answer, the
function returns the error. The post-commit hook already has a branch
for that error, and that branch writes a warning and continues, so hooks
stay non-blocking.

**The benchmark holds both strategies.** The benchmark lives in the git
package as a test-only file. It builds a repository with a large file
count, then measures two strategies over the same commit:

- The two-commit comparison that the code uses today.
- The `git diff-tree` plumbing call that issue #30 proposed.

The `git diff-tree` call appears only in that file. Production code
never gains a second strategy. The benchmark is a measurement tool, and
continuous integration does not gate on its numbers.

**The file count is a benchmark parameter.** The benchmark runs at a
size that matches the claim under test, and the size appears in the
benchmark name so a reader knows what was measured.

## Testing Decisions

**What makes a good test here.** A test drives a public function of the
package and asserts the value that function returns. It does not assert
which git subcommand ran, which flags were passed, or which branch of
the helper executed. A test that survives a rewrite of the internals is
a good test. A test that fails when the internals move is not.

**Every module is tested.** The three commit-diff functions and the
attribution function each get tests. This is not negotiated.

**Each function gets the same four cases.** Every commit-diff function
is tested against a first commit, a regular commit, a merge commit, and
a commit that adds a binary file. The first-commit case is the guard for
this defect. The merge-commit case is the guard against the regression
found in pull request #653.

**Attribution is tested for both actors and for failure.** The tests
cover an agent-active first commit, a human first commit, and a commit
hash that git cannot resolve. The failure case asserts that the function
returns an error, because the silent zero is the defect under repair.

**Prior art.** The git package already holds a test that builds a real
repository in a temporary directory, runs git through a local helper
closure, and drives table-driven subtests on the standard library alone.
The new tests follow it. The repository uses no external test framework,
and none is added.

**Named gap: the tests change the process working directory.** The three
commit-diff functions read the process working directory, so a test must
change directory to reach a temporary repository. This is a wart in the
existing interface. This work does not repair it, because a repository
path parameter would ripple to every caller. The gap is named here so it
is visible rather than silent.

**Named gap: no test asserts hook latency.** The benchmark measures, it
does not assert. No test fails when the post-commit path gets slower.
The benchmark exists so a future claim can be checked by a person, not
so continuous integration can block a merge.

## Out of Scope

- **Pull request #653.** It stays open. The decision to close it belongs
  to the operator. This work does not close it, does not comment on it,
  and does not push to its branch.
- **Issue #30.** It is already closed. This work does not reopen it.
- **The blanket `git diff-tree` replacement.** It is measured as level
  on a repository with one thousand files, and roughly thirty to forty
  percent slower on a repository with ten thousand files. It also zeroes
  merge attribution. It is not adopted.
- **A repository path parameter on the commit-diff functions.** The
  interface wart is named in Testing Decisions and left alone.
- **The post-commit hook's fallback to one hundred percent on an
  attribution error.** That behaviour is odd, and it is untouched here.
- **Attribution by complexity.** Attribution stays binary, at zero or
  one hundred percent, as it is today.
- **Any release, tag, or version pin bump.**
- **Other repositories.** Only the CLI repository changes.
- **A correction to the premise-gate document.** The correction is
  recorded in Further Notes below. Whether to edit that document is a
  separate decision.

## Further Notes

**The premise-gate document states the wrong cause.** It says the
well-known empty tree object "is not present in a fresh repository — nor
in this one". That is false. A comparison against
`4b825dc642cb6eb9a060e54bf8d69288fbee4904` succeeds in a repository
created seconds earlier, because git special-cases that object. The
cause is the corrupted constant alone.

This matters more than a footnote. The premise-gate document exists
because a proposal asserted something about this repository that nobody
checked. Its own account of this defect is an assertion that nobody
checked. The gate it describes would have caught it.

**Pull request #653 regresses merge commits.** It routes the line counts
and the changed file list through `git diff-tree --no-commit-id -r
--root`. On a merge commit, the current code reports the lines that the
merge brings in against its first parent. The proposed code reports
nothing. Merge attribution drops to zero, silently. Issue #30's own
acceptance criteria required "identical results to the previous
implementation for regular commits, merges, and initial commits", so the
pull request fails a criterion that it set for itself.

**The first-commit defect was found by accident.** It is unrelated to
the stated goal of issue #30, and it is the only real defect in that
episode. The salvage keeps that fix and nothing else.

**The benchmark is the durable part.** The line-count fix stops a
present defect. The benchmark stops the next false speed claim, and this
path has already attracted one. Process startup is roughly two thirds of
the total in both strategies, which is the more useful finding: the
comparison strategy was never the bottleneck here.
