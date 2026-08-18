# 05 — The speed claim is measurable

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — First commit records its lines](./01-first-commit-records-lines.md)

## What to build

A benchmark measures the post-commit comparison strategy, so the next
speed claim about this path meets a number.

Issue #30 said the post-commit hook performed a full tree walk that was
O(N) in the number of tracked files, and that a `git diff-tree`
replacement would cut hook latency. The hook never performed a tree
walk. It always ran a two-commit `git diff`, which is a tree-to-tree
comparison. The claim came from a sibling product's changelog. The
repository holds no benchmark, so nobody could check the claim, and a
pull request was built on it four months later.

This slice adds a benchmark in the git package that runs two strategies
over the same commit in the same generated repository:

- The two-commit comparison the code uses today.
- The `git diff-tree --no-commit-id -r --root` plumbing call that issue
  #30 proposed.

**The `git diff-tree` call lives only in the benchmark file.**
Production code keeps exactly one strategy. Do not add a production
helper for the rejected strategy, and do not export it.

The benchmark builds a repository with a large file count. The file
count appears in the benchmark name, so a reader knows what was
measured. Continuous integration does not gate on the numbers. The
benchmark is a measurement tool for a person.

When the benchmark runs, record the numbers in this issue file under the
Handoff block below.

The expected result, from the premise-gate document's own measurements:
the replacement is level on a repository with one thousand files, and
roughly thirty to forty percent slower on a repository with ten thousand
files. Git process startup is roughly two thirds of the total in both
strategies. Report what you actually measure. If the numbers disagree
with that expectation, say so and do not adjust them.

## User stories covered

18, 19, 20, 21, 22, 23, 24, 28, 30

## Acceptance criteria

- [x] The git package holds a benchmark that measures the current
      comparison strategy over a generated repository.
- [x] The same benchmark file measures the `git diff-tree` alternative
      over the same commit in the same repository.
- [x] The `git diff-tree` invocation appears only in the benchmark file
      and in no production file.
- [x] The generated repository's file count appears in the benchmark
      name.
- [x] The benchmark builds its repository under `t.TempDir()` or the
      benchmark equivalent, so it leaves nothing behind.
- [x] `go test -bench` runs both strategies and reports a number for
      each.
- [x] `make test` passes and does not run the benchmark by default.
- [x] The measured numbers are recorded in the Handoff block in this
      file.

## Modules touched

- The git package benchmark file (test-only).

## Test prior art

- `internal/git/*_test.go` — the pattern that builds a real repository
  in a temporary directory and runs git through a local `run` closure.
- The repository holds no benchmark today, so there is no prior art for
  the benchmark shape itself. Follow the standard library `testing`
  package's `Benchmark` convention and keep the repository-building
  helper out of the timed loop.
- The house pattern is the standard library `testing` package alone. Do
  not add a benchmark framework.

## Out of scope

- Any production change to the comparison strategy. The alternative is
  measured, not adopted.
- A test that fails when the post-commit path gets slower. The PRD names
  this gap: the benchmark measures, it does not assert.
- The root-commit fix in the three commit-diff functions. Slices
  [01](./01-first-commit-records-lines.md),
  [02](./02-first-checkpoint-lists-files.md) and
  [03](./03-first-checkpoint-stores-diff.md) cover them.
- Closing pull request #653. See the Handoff below.

## Handoff

When this slice's acceptance criteria all pass, BEFORE flipping its row
in `issues.md`, post the following block to the user verbatim:

- **URL / artefact to visit**:
  - The benchmark output from `go test -bench` in the git package, with
    both strategies side by side.
  - https://github.com/partio-io/cli/pull/653
- **Action required**: Read the two benchmark numbers. Pull request #653
  replaces the current strategy with the slower one, and it also drops
  merge-commit attribution to zero. This work does not close it. Decide
  whether to close pull request #653, and close it yourself.
- **Where to record the decision**:
  - In this file, under this Handoff block, paste the measured numbers
    so the next proposal about this path meets a recorded figure, AND
  - In `issues.md`, add a one-line note next to this slice's row stating
    whether pull request #653 was closed.

There is no next slice, so no slice is blocked on this. The handoff
exists because the decision belongs to the operator and would otherwise
go unrecorded.

**Measured numbers**: Recorded 2026-08-17.

Command: `go test -run '^$' -bench BenchmarkCommitDiffStrategies
-benchtime 100x -count 5 ./internal/git/`

Machine: AMD Ryzen AI MAX+ 395, linux/amd64, 32 threads. git 2.55.0, go
1.26.6. Each figure is the median of 5 runs, in milliseconds for one
call. The measured commit changes one file in a repository that holds
the stated file count.

| Measurement | 1000 files | 10000 files |
|---|---|---|
| Current strategy, `DiffNumstat` (parent lookup + `git diff`) | 1.34 | 1.74 |
| The `git diff` command alone, revisions already known | 0.74 | 1.17 |
| Proposed `git diff-tree --no-commit-id -r --root --numstat` | 0.74 | 1.47 |

**Read the third row against the second row, not the first.** The first
row starts two git processes, because the current code asks git for the
parent commit first. The third row starts one. A git process costs about
0.6 ms on this machine, and that fixed cost, not the tree size, is the
whole gap between row one and row three.

Command against command, rows two and three:

- 1000 files: the proposal is level. It is 0.3 percent slower, which is
  inside the noise.
- 10000 files: the proposal is 26 percent slower.

This agrees with the premise-gate document's direction. The document
expected 30 to 40 percent slower at 10000 files; the measurement here is
26 percent. The proposal is slower at scale, and it is never faster.

Strategy against strategy, rows one and three, the proposal looks 45
percent faster at 1000 files and 16 percent faster at 10000 files. That
number is real but it does not measure the tree comparison. It measures
one git process against two. Slice 01 added the parent lookup on
purpose, to stop the code from guessing the parent. A future proposal
that cites the faster number must say which of the two it means.

Process startup share: about 80 percent of a single call at 1000 files
and about 50 percent at 10000 files. The premise-gate document said
roughly two thirds. The measured share is not constant; it falls as the
repository grows.

**Issue #30's premise is not supported.** The claim was a full tree walk
that is O(N) in tracked files. A ten times larger repository raises the
current call from 1.34 ms to 1.74 ms, which is 30 percent, not ten
times. The cost is dominated by process startup.
