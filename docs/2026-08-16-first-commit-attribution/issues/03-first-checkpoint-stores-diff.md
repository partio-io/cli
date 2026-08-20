# 03 — First checkpoint stores its diff

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — First commit records its lines](./01-first-commit-records-lines.md)

## What to build

A checkpoint for a first commit holds the unified diff of that commit.
Today the unified-diff function compares a commit against its parent, so
it fails on a first commit. The post-commit hook calls it inside an
`err == nil` check, so the hook silently stores a checkpoint with no
diff at all. The user's first checkpoint holds the agent session and
none of the code.

The function has the same root-commit defect that slice 01 repaired in
the line-count function, and slice 01 already built the fix. This slice
routes the unified-diff function through the same commit-range helper.

The function keeps its current signature, so no caller changes. It
builds its own flags and appends the two revisions from the helper.

**Merge behaviour must not change.** A merge commit has a parent, so it
takes the unchanged path and is compared against its first parent. Do
NOT add the `--root` flag to production code.

## User stories covered

3, 4, 17, 29, 30, 32, 33, 34

## Acceptance criteria

- [x] The unified-diff function uses the commit-range helper from slice
      01 and holds no revision logic of its own.
- [x] The unified-diff function keeps its current signature.
- [x] The unified-diff function on a first commit returns a diff that
      contains the added content.
- [x] The unified-diff function on a regular commit returns the same
      diff it returns today.
- [x] The unified-diff function on a merge commit returns the diff
      against its first parent, the same diff it returns today.
- [x] The unified-diff function on an empty commit returns an empty
      result and no error.
- [x] A checkpoint created for a first commit holds a non-empty diff.
- [x] `make test` passes.

## Modules touched

- The git package unified-diff function (`git.Diff`).

## Test prior art

The git package already holds a test that builds a real repository in a
temporary directory, runs git through a local `run` closure that calls
`t.Fatalf` on failure, and drives table-driven subtests. Slice 01 added
tests in the same package under the same pattern — mirror them.

- `internal/git/*_test.go` — the repository-building pattern.
- `internal/hooks/*_test.go` — prior art for a test that drives the
  post-commit path, if the checkpoint criterion needs one.
- The house pattern is `t.TempDir()` for filesystem isolation, table-
  driven subtests, and the standard library `testing` package alone.
- The function reads the process working directory, so a test must
  change directory into the temporary repository with `t.Chdir`.

## Out of scope

- The line-count function and the shared helper. Slice
  [01](./01-first-commit-records-lines.md) covers them.
- The changed-file function. Slice
  [02](./02-first-checkpoint-lists-files.md) covers it.
- The checkpoint storage format. This slice changes what the hook can
  fetch, not how a checkpoint is written.
- A repository path parameter on the commit-diff functions.
- Pull request #653 and issue #30. Do not close, comment on, or push to
  either.
