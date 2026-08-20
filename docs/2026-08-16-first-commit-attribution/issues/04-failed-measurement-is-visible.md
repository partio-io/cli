# 04 — A failed measurement is visible

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — First commit records its lines](./01-first-commit-records-lines.md)

## What to build

When git cannot measure a commit, the user learns about it. Today the
attribution function returns an empty result and a nil error, so a
failed measurement and an empty commit look identical. The post-commit
hook has a branch for an attribution error, and that branch never runs.

This is the behaviour that hid the first-commit defect for four months.
The defect was found by accident, and it was found by accident because
nothing ever reported it.

This slice makes the attribution function return the error when git
cannot answer. The post-commit hook's existing error branch then runs,
writes a warning, and continues.

**Hooks must stay non-blocking.** The repository's convention is that a
hook logs a warning and does not fail a git operation on a non-critical
error. A commit must still succeed when attribution fails. Do not turn
this error into a hook failure.

The hook's existing error branch sets the result to one hundred percent
agent lines. That behaviour is odd, and the PRD puts it out of scope.
Leave it as it is.

## User stories covered

14, 15, 29, 30, 32

## Acceptance criteria

- [x] The attribution function returns the error from git when git
      cannot measure the commit.
- [x] The attribution function no longer returns an empty result with a
      nil error on a git failure.
- [x] The attribution function on an empty commit still returns zero
      lines and a nil error, so a real zero and a failed measurement
      stay distinguishable.
- [x] A commit hash that git cannot resolve produces an error from the
      attribution function.
- [x] The post-commit hook writes a warning when attribution fails.
- [x] The post-commit hook does not fail the git operation when
      attribution fails.
- [x] `make test` passes.

## Modules touched

- The attribution calculate function (`attribution.Calculate`).
- The post-commit hook, for the warning path only.

## Test prior art

- `internal/hooks/*_test.go` — prior art for tests that drive a hook
  against a real repository in a temporary directory.
- `internal/git/*_test.go` — the repository-building pattern, and the
  tests slice 01 added.
- The house pattern is `t.TempDir()` for filesystem isolation,
  `t.Setenv()` for environment variables, table-driven subtests, and the
  standard library `testing` package alone.

## Out of scope

- The root-commit fix in any of the three commit-diff functions. Slices
  [01](./01-first-commit-records-lines.md),
  [02](./02-first-checkpoint-lists-files.md) and
  [03](./03-first-checkpoint-stores-diff.md) cover them.
- The hook's fallback to one hundred percent agent lines on an
  attribution error. The PRD names it and leaves it alone.
- Any change that makes a hook block a git operation.
- Pull request #653 and issue #30. Do not close, comment on, or push to
  either.
