# 02 — First checkpoint lists its files

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — First commit records its lines](./01-first-commit-records-lines.md)

## What to build

Partio lists the files that a first commit touched. Today the changed-
file function compares a commit against its parent, so it fails on a
first commit and returns an empty list.

The function has the same root-commit defect that slice 01 repaired in
the line-count function, and slice 01 already built the fix. This slice
routes the changed-file function through the same commit-range helper.

The function keeps its current signature, so no caller changes. It
builds its own flags and appends the two revisions from the helper.

**Merge behaviour must not change.** A merge commit has a parent, so it
takes the unchanged path and is compared against its first parent. Do
NOT add the `--root` flag to production code.

The current caller of this function is the post-commit hook's debug
logging, which compares the commit's file list against the agent session
content paths. That caller is behind a debug-level check, so the visible
outcome of this slice is a correct file list at debug level on a first
commit, plus a function that later callers can trust.

## User stories covered

2, 4, 17, 29, 30, 32, 33, 34

## Acceptance criteria

- [x] The changed-file function uses the commit-range helper from slice
      01 and holds no revision logic of its own.
- [x] The changed-file function keeps its current signature.
- [x] The changed-file function on a first commit returns every path
      that commit added.
- [x] The changed-file function on a regular commit returns the same
      paths it returns today.
- [x] The changed-file function on a merge commit returns the paths the
      merge brings in against its first parent, the same paths it
      returns today.
- [x] The changed-file function on a first commit that adds a binary
      file includes that file's path.
- [x] The changed-file function on an empty commit returns an empty
      result and no error.
- [x] `make test` passes.

## Modules touched

- The git package changed-file function (`DiffNameOnly`).

## Test prior art

The git package already holds a test that builds a real repository in a
temporary directory, runs git through a local `run` closure that calls
`t.Fatalf` on failure, and drives table-driven subtests. Slice 01 added
tests in the same package under the same pattern — mirror them.

- `internal/git/*_test.go` — the repository-building pattern.
- The house pattern is `t.TempDir()` for filesystem isolation, table-
  driven subtests, and the standard library `testing` package alone.
- The function reads the process working directory, so a test must
  change directory into the temporary repository with `t.Chdir`.

## Out of scope

- The line-count function and the shared helper. Slice
  [01](./01-first-commit-records-lines.md) covers them.
- The unified-diff function. Slice
  [03](./03-first-checkpoint-stores-diff.md) covers it.
- Any change to the post-commit hook's debug logging. This slice repairs
  the function the hook calls, not the hook.
- A repository path parameter on the commit-diff functions.
- Pull request #653 and issue #30. Do not close, comment on, or push to
  either.
