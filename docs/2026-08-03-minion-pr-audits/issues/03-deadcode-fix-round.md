# 03 — Dead-code fix round

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Dead-code audit, read-only](./02-deadcode-audit-readonly.md)

## What to build

The bounded repair loop on top of issue 02: when the audit's verdict
is `fail`, run exactly one fix session against the findings, commit
the result to the PR branch with the `audit:` subject prefix, push,
and re-audit once with a fresh session. The re-audit's verdict is
final — green if the branch came back clean (the operator then reviews
a PR that is already fixed), red plus the surviving-findings comment
if not. Never a second fix attempt, regardless of outcome.

The loop guard installed in issue 02 now earns its keep: the audit's
own push fires a new workflow run, which must recognize the `audit:`
head commit and exit early instead of auditing itself.

## User stories covered

2, 3, 16, 25 (comment-convergence half; proposal half lives in 04).

## Acceptance criteria

- [ ] Fixable staged PR: planted dead code is repaired by a commit
      with the `audit:` prefix on the PR branch, the re-audit passes,
      the check ends green, and the PR carries at most the in-place
      audit comment (updated to reflect resolution or absent).
- [ ] The workflow run triggered by the audit's own push exits early
      via the guard — observed on the staged PR, no self-audit.
- [ ] Unfixable staged PR (a finding whose fix is deliberately out of
      a session's reach, e.g. requiring a design decision): exactly
      one fix attempt, then red with the surviving findings in the
      single comment.
- [ ] Re-running the workflow on an unchanged PR converges: same
      verdict, no duplicate comments, no second fix commit.
- [ ] Fix commits pass the repo's own checks (lint from issue 01
      included) before pushing — a fix that breaks the build is a
      failed fix round, red.
- [ ] Dry-run still touches nothing.

## Modules touched

- deadcode-audit program (fix-round instructions and re-audit)
- deadcode-audit workflow (fix/re-audit steps, guard consumption)
- audit verdict gate (unchanged, consumed)

## Test prior art

- Issue 02's staged-scratch-PR pattern — extend the same staging with
  the fixable and unfixable variants.
- Check-before-push discipline: the minions runtime's per-slice
  check-gate behavior (checks run, retry once, fail the slice) is the
  conceptual model for "a fix that fails checks is a failed round".

## Out of scope

- Multi-round repair, escalation, or fix attempts on human PRs.
- E2e-need findings — issue 04.
