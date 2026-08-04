# 04 — E2e-need audit

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Dead-code audit, read-only](./02-deadcode-audit-readonly.md)

## What to build

The second audit action, reusing issue 02's trigger gating and verdict
contract: a program that judges whether the PR's change warrants an
end-to-end test, and when it does, feeds that work back into the
pipeline as a `minion-proposal` issue instead of blocking the PR.

The judgment exemplar (from the PR #586 review): a fix whose headline
acceptance criterion — "two rapid commits in one live session both get
trailers" — was verified only via unit tests of the two halves of the
mechanism, with nothing driving both hooks through the real two-commit
flow. That gap is what this audit exists to name.

A finding becomes one proposal per distinct gap, matching the propose
program's issue format: kebab-case id, a program file committed and
pushed the way propose already does it, and an issue whose body
carries acceptance criteria concrete enough for a fresh minion session
to build the test from the issue alone. Before filing, the program
searches open proposals and skips ids that already exist.

Findings never turn this check red — a completed run with findings is
a `pass` verdict plus filed proposals. Only a missing or malformed
verdict (a crashed session) fails the workflow.

## User stories covered

8, 9, 10, 11, 15, 17, 18, 19, 20, 21, 25 (proposal-dedup half).

## Acceptance criteria

- [x] Same trigger surface as the dead-code audit: minion label
      required, `no-audit` skip, `audit:` head-commit guard,
      concurrency cancel, same runner, same minions pin, dry-run via
      `workflow_dispatch` that touches nothing.
- [x] Staged PR that plainly deserves e2e coverage: exactly one
      `minion-proposal` issue filed, body prefixed `Minion audit —`,
      with buildable acceptance criteria and the program-file
      reference; check green.
- [x] Re-run on the same staged PR: no duplicate proposal, still
      green.
- [x] Staged cosmetic PR (comment/docs-only change): no proposal, no
      comment, green.
- [x] Fixture-verdict proof that a missing or malformed verdict still
      fails this workflow despite findings never doing so.
- [x] Staged artifacts cleaned up afterwards (scratch PRs closed,
      scratch proposals closed and labeled `do-not-build`).

## Modules touched

- e2e-audit program
- e2e-audit workflow
- audit verdict gate (reused as-is)

## Test prior art

- Issue 02's workflow and staging patterns.
- Proposal filing and dedup: the propose program under
  `.minions/programs/` — reuse its id-generation, existence-check, and
  issue-format conventions verbatim.

## Out of scope

- Building the e2e tests themselves — that is future minion work,
  entering through the proposals this audit files.
- Blocking merges on missing e2e coverage.
