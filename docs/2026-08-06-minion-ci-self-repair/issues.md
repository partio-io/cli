# Issues — 2026-08-06-minion-ci-self-repair

Source: [prd.md](./prd.md)

| Done | # | Title | Blocked by |
|------|---|-------|------------|
| [x]  | 1 | [Unblock the dead-code fixer](./issues/01-unblock-deadcode-fixer.md) | None |
| [x]  | 2 | [Count repair rounds instead of forbidding them](./issues/02-count-repair-rounds.md) | [#1](./issues/01-unblock-deadcode-fixer.md) |
| [x]  | 3 | [Say when it gave up](./issues/03-rounds-exhausted-comment.md) | [#2](./issues/02-count-repair-rounds.md) |
| [x]  | 4 | [One shared apply-prove-push](./issues/04-shared-patch-apply.md) | [#2](./issues/02-count-repair-rounds.md) |
| [x]  | 5 | [Lint and test as a PR check](./issues/05-lint-test-pr-check.md) | None |
| [ ]  | 6 | [Repair failing lint and tests](./issues/06-repair-failing-checks.md) | [#4](./issues/04-shared-patch-apply.md), [#5](./issues/05-lint-test-pr-check.md) |
| [x]  | 7 | [Stop the gates colliding](./issues/07-gate-concurrency-and-retry.md) | [#6](./issues/06-repair-failing-checks.md) |

## Notes

Every slice lands on `main` through a pull request that jcleira merges;
no slice merges itself. Where a slice's proof requires the change to be
live on `main` — slice 1's re-run of PR #622 is the clearest case —
that step is called out in the slice's own acceptance criteria as a
post-merge verification.

Slice 6 is built and its row stays `[ ]` for one reason: its last
criterion is a post-merge verification that no working tree can show.
Do not implement it again. Merge it, watch the first minion pull request
that breaks its own test, then check the row.

Slices 1 and 5 have no blockers and can start immediately.
