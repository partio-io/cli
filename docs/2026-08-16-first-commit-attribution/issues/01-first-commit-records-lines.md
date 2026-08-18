# 01 — First commit records its lines

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

Partio counts the added lines in the first commit of a repository. A
user who installs Partio and commits gets a checkpoint that reports the
real line total, not zero.

Today the line-count function compares a commit against its parent. A
first commit has no parent, so git answers `fatal: ambiguous argument
'HEAD~1'`. The attribution code then falls back to a comparison against
git's empty tree object, and the constant it uses is corrupted:

- The code holds `4b825dc642cb6eb9a060e54bf899d69f82cf7ee2`.
- The real empty tree object is
  `4b825dc642cb6eb9a060e54bf8d69288fbee4904`.

Git answers the fallback with `fatal: bad object`. The attribution code
catches that error, returns an empty result, and reports no error, so
the first commit records zero lines.

This slice adds one named constant for the empty tree object and one
unexported helper in the git package that decides how to identify a
commit's content. The helper takes a commit hash and returns the two
revision arguments to compare:

- A commit with a parent returns the parent revision and the commit
  revision. This is today's behaviour, unchanged.
- A commit with no parent returns the empty tree constant and the commit
  revision.

The line-count function then builds its own flags and appends the two
revisions from the helper. Its signature does not change, so no caller
changes. The attribution function drops its private copy of the empty
tree fallback, which the helper makes dead.

**Merge behaviour must not change.** A merge commit has a parent, so it
takes the unchanged path and is compared against its first parent. Do
NOT add the `--root` flag to production code. Measured directly, `git
diff-tree --no-commit-id -r --root --numstat` returns nothing at all for
a merge commit, while today's comparison returns the lines the merge
brings in. Pull request #653 made that mistake, and this slice exists in
part to avoid repeating it.

## User stories covered

1, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 16, 17, 29, 30, 31, 32, 33, 34

## Acceptance criteria

- [x] The git package holds one exported constant for git's empty tree
      object id, with a comment that states what it is.
- [x] The corrupted literal `4b825dc642cb6eb9a060e54bf899d69f82cf7ee2`
      appears nowhere in the repository.
- [x] An unexported helper in the git package takes a commit hash and
      returns the two revision arguments that identify that commit's
      content.
- [x] The helper detects the parent through git, not through a guess,
      and an error that is not the absence of a parent propagates to the
      caller.
- [x] The line-count function keeps its current signature.
- [x] Attribution on a first commit that adds text files reports the
      real added-line total.
- [x] Attribution on a first commit with an agent active reports one
      hundred percent agent lines and the real total.
- [x] Attribution on a first commit with no agent active reports zero
      percent agent lines and the real total.
- [x] Attribution on a first commit that adds a binary file skips that
      file and still counts the text files.
- [x] Attribution on a first empty commit reports zero lines and no
      error, so a real zero differs from a failed measurement.
- [x] Attribution on a regular commit reports the same total it reports
      today.
- [x] Attribution on a merge commit reports the lines the merge brings
      in against its first parent, the same total it reports today.
- [x] The attribution function no longer holds its own empty tree
      fallback.
- [x] `make test` passes.

## Modules touched

- The git package empty tree constant.
- The git package commit-range helper.
- The git package line-count function (`DiffNumstat`).
- The attribution calculate function (`attribution.Calculate`).

## Test prior art

The git package already holds a test that builds a real repository in a
temporary directory, runs git through a local `run` closure that calls
`t.Fatalf` on failure, and drives table-driven subtests. Follow it.

- `internal/git/*_test.go` — the repository-building pattern.
- The house pattern is `t.TempDir()` for filesystem isolation, table-
  driven subtests, and the standard library `testing` package alone.
- The repository uses no external test framework. Do not add one.

The three commit-diff functions read the process working directory, so a
test must change directory into the temporary repository with `t.Chdir`.
That is a known wart in the interface and this slice does not repair it.

## Out of scope

- The changed-file function. Slice
  [02](./02-first-checkpoint-lists-files.md) covers it.
- The unified-diff function. Slice
  [03](./03-first-checkpoint-stores-diff.md) covers it.
- Error propagation out of the attribution function when git genuinely
  fails. Slice [04](./04-failed-measurement-is-visible.md) covers it.
  In this slice the attribution function keeps its current behaviour on
  a real failure.
- The benchmark. Slice [05](./05-speed-claim-is-measurable.md) covers
  it.
- A repository path parameter on the commit-diff functions. The PRD
  names the interface wart and leaves it alone.
- Pull request #653 and issue #30. Do not close, comment on, or push to
  either.
