# 05 — Lint and test as a PR check

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

There is no `make test` or `make lint` check on pull requests. Both run
inside the minion session and inside the audit's apply step, but
neither produces a check, so a broken build never turns anything red.
It is found by reading the diff, or after merging.

Add the gate. It runs on every pull request — hand-written ones
included — and reports red or green like the audits do. This slice
delivers signal only; repair arrives in slice
[06](./06-repair-failing-checks.md).

The design point that makes the rest of the work cheap: the gate emits
the *same verdict artifact the audits already emit* — a status plus a
list of findings, each with a location and the reasoning behind it.
Because the shape matches, the existing verdict gate command, the pull
request comment surface, the round accounting from slice
[02](./02-count-repair-rounds.md), and the patch applier from slice
[04](./04-shared-patch-apply.md) all work on lint and test failures
without knowing what produced them.

Failures must carry enough reasoning for a later repair session to act
without re-running anything itself: which package or linter failed,
where, and what the failure said.

## User stories covered

PRD user stories 7, 8, 9, 23, 29.

## Acceptance criteria

- [ ] `make lint` and `make test` run as a check on every pull request,
      including pull requests with no minion label.
- [ ] Failures are converted into the same verdict shape the audits
      emit, one finding per failing package or linter finding.
- [ ] Each finding carries a location and reasoning sufficient to act on
      without re-running the command.
- [ ] A clean run yields a pass verdict and a green check.
- [ ] The check reports through the existing pull request comment
      surface, upserting one comment rather than appending per run.
- [ ] A missing or malformed verdict fails the check closed, matching
      how the audits behave — a gate must never pass by producing
      nothing.
- [ ] No repair is attempted and no commit is pushed by this slice, on
      any pull request.
- [ ] The verdict conversion is covered by tests over captured command
      output fixtures: a clean run, a failing test, a lint finding, and
      unparseable output — asserting the resulting status and findings.
- [ ] The work requires no minions release and no version-pin bump.

## Modules touched

`internal/checksverdict` (new) and the new checks workflow from the
PRD's Implementation Decisions ("Workflows"). Reuses the existing
verdict gate command and `internal/auditgate`'s comment surface
unchanged.

## Test prior art

`internal/auditgate/*_test.go` for verdict loading and gate outcomes,
including the fail-closed paths for a missing or malformed verdict —
this slice must not regress that behavior for its own verdicts.

Command-output fixtures belong in `testdata/` per Go convention.
Table-driven, standard library `testing` only, `t.TempDir()` for
filesystem isolation.

## Out of scope

- Repairing anything. The gate reports; slice
  [06](./06-repair-failing-checks.md) makes it fix.
- Pushing commits to any branch, including minion branches.
- Any quality signal beyond lint and test. Others may earn a gate later
  and would inherit the repair contract rather than extend this scope.
- Concurrency with the audit gates, added in slice
  [07](./07-gate-concurrency-and-retry.md).
