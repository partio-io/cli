# 06 — Repair failing lint and tests

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — One shared apply-prove-push](./04-shared-patch-apply.md), [05 — Lint and test as a PR check](./05-lint-test-pr-check.md)

## What to build

The lint/test gate reports failures but does nothing about them. Give
it the repair round the dead-code audit has, using the machinery the
previous slices built: round accounting decides whether another attempt
is allowed, a repair session produces a patch, and the shared applier
proves and pushes it.

Repair is attempted only on pull requests that carry the minion label
and whose head branch lives in this repository. Hand-written branches
get the red/green signal and nothing else — no bot commits are pushed
to work the operator wrote.

The repair session's boundary is what makes this safe. It may modify
implementation code freely. It may modify a test **only when that test
was added by the pull request under repair** — a minion's own new test
being wrong is the common case and should not become the operator's
queue. A test that predates the pull request is out of bounds, and a
failure in one is a surviving finding that goes red. This is the one
rule that prevents the classic failure where a bot reaches green by
deleting an assertion.

The session follows the same contract as the dead-code fixer: produce a
patch, never commit, never push, never comment, and write the skip
marker as its final act on every path.

## User stories covered

PRD user stories 10, 11, 12, 27.

## Acceptance criteria

- [x] A failing lint/test verdict on a minion-labeled pull request
      triggers a repair round.
- [x] Repair is skipped for pull requests without the minion label, and
      no commit is pushed to them.
- [x] Repair is skipped for pull requests whose head branch lives in a
      fork, with a clear reason and no push attempt.
- [x] The repair session may modify implementation code and may modify
      tests the same pull request added.
- [x] The repair session does not modify tests that predate the pull
      request; a failure in one survives as a finding and the check goes
      red.
- [x] The session produces a patch only, and runs no git write command
      on any path.
- [x] Rounds are capped at three for this check, counted independently
      of the dead-code check's rounds.
- [x] Repairs are applied, proven, and pushed through the shared applier
      from slice [04](./04-shared-patch-apply.md), not through new
      inline shell.
- [ ] Post-merge verification: a pull request whose failing test the
      minion itself wrote goes green without human edits. This one needs
      the workflow live on `main` and a minion pull request that breaks
      its own test; it cannot be shown from the working tree.
- [x] Rounds exhausted with findings surviving produces the give-up
      comment from slice [03](./03-rounds-exhausted-comment.md), naming
      the survivors and the round count.

## Modules touched

The checks fix program (new) from the PRD's Implementation Decisions,
plus the checks workflow (modified). Consumes `internal/repairround`,
`internal/patchapply`, `internal/checksverdict`, and `internal/auditgate`
— no new Go package is introduced by this slice.

## Test prior art

The repair program itself is model-driven and not unit-testable; the
PRD names this in Testing Decisions. Its contract is enforced
downstream — a session that produces a bad patch fails the applier's
`make lint && make test` and pushes nothing.

The Go-side behavior this slice wires together is already covered by
the tests from slices [02](./02-count-repair-rounds.md),
[04](./04-shared-patch-apply.md), and
[05](./05-lint-test-pr-check.md). Any new decision logic added here —
label and fork gating in particular — gets its own table-driven tests
in the style of `internal/auditgate/*_test.go`.

Distinguishing PR-added tests from pre-existing ones is decidable from
the pull request diff; that determination is Go-side logic and must be
tested, not left to the session's judgment alone.

## Out of scope

- Widening the gate beyond lint and test.
- Repairing hand-written or fork pull requests.
- Rebuilding a pull request that exhausts its rounds — it stops and
  reports.
- Concurrency between this gate and the audit gates, which slice
  [07](./07-gate-concurrency-and-retry.md) resolves. Until that lands,
  two gates repairing the same branch at once remains possible.
