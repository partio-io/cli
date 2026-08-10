# 02 — Count repair rounds instead of forbidding them

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — Unblock the dead-code fixer](./01-unblock-deadcode-fixer.md)

## What to build

Today a repaired pull request can never be repaired again. The audit
workflow's loop guard reads the head commit's subject line and skips
the entire run when it looks like a repair commit. That stops runaway
loops, but it also means a repair that lands one edit short of green is
indistinguishable from a repair that was never attempted.

Replace the guard with accounting. Before an audit runs, ask how many
repair rounds for *this* check already sit on the pull request's
branch. Under the cap, the run proceeds and knows which round it is;
at the cap, the run stops without auditing again.

The cap is three rounds per check, counted independently. A dead-code
repair must never consume the budget a failing test would need, so
repair commits have to carry a marker naming which check produced
them, and the count for one check must ignore commits belonging to
another. Commits authored by a human, and commits whose subject merely
resembles a repair marker without being one, must not be counted.

This slice introduces the accounting as a tested package and wires the
dead-code audit workflow to it. Once landed, a dead-code finding that
survives its first repair gets a second and a third attempt, and the
fourth run declines.

## User stories covered

PRD user stories 4, 5, 6, 19, 25.

## Acceptance criteria

- [x] A new package owns the decision "may another repair round start
      for this check, and which round is it?", taking the branch's
      commit subjects, the check name, and the cap as input.
- [x] Repair commits are marked so that the check that produced them is
      recoverable from the commit subject alone.
- [x] Counting for one check ignores repair commits belonging to another
      check.
- [x] The cap is three rounds per check, and the package reports the
      round number so a caller can distinguish the first attempt from
      the last.
- [x] Human-authored commits and near-miss subjects are not counted as
      repair rounds.
- [x] The dead-code audit workflow consults the package instead of its
      inline shell skip guard, and stops auditing once the budget for
      that check is spent.
- [x] A pull request whose first repair leaves findings receives a
      second repair attempt rather than going red immediately.
- [x] The package is covered by table-driven tests over commit-subject
      lists: no prior rounds, one, at the cap, over the cap, another
      check's rounds interleaved, human commits interleaved, an empty
      branch, and a subject that resembles a marker without being one.

## Modules touched

`internal/repairround` (new), and the dead-code audit workflow from the
PRD's Implementation Decisions ("Workflows").

## Test prior art

`internal/*/[a-z]*_test.go` throughout the repo: table-driven, standard
library `testing` only, no external assertion frameworks. `t.TempDir()`
for filesystem isolation and `t.Setenv()` for environment variables.
`internal/auditgate/*_test.go` is the closest sibling in both subject
matter and shape.

The repository convention of one primary concern per file applies to
the new package.

## Out of scope

- What happens after the budget is spent. The pull request simply stops
  being audited; slice [03](./03-rounds-exhausted-comment.md) adds the
  comment that says so.
- Moving the apply-prove-push shell into Go. Slice
  [04](./04-shared-patch-apply.md) does that.
- The lint/test gate, which does not exist yet. Slice
  [05](./05-lint-test-pr-check.md) adds it, and slice
  [06](./06-repair-failing-checks.md) gives it repair rounds using this
  package.
- Concurrency between gates. Slice
  [07](./07-gate-concurrency-and-retry.md) covers it.
