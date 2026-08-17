# Issues — 2026-08-13-proposal-premise-gate

Source: [prd.md](./prd.md)

| Done | # | Title | Blocked by |
|------|---|-------|------------|
| [x]  | 1 | [Program shape check](./issues/01-program-shape-check.md) | None |
| [x]  | 2 | [Propose instructions reach the model](./issues/02-propose-instructions-reach-model.md) | [#1](./issues/01-program-shape-check.md) |
| [x]  | 3 | [Proposals carry a premise](./issues/03-proposals-carry-premise.md) | [#2](./issues/02-propose-instructions-reach-model.md) |
| [x]  | 4 | [The proposer checks before it files](./issues/04-proposer-checks-before-filing.md) | [#3](./issues/03-proposals-carry-premise.md) |
| [x]  | 5 | [Rejected ideas leave a record](./issues/05-rejection-log.md) | [#4](./issues/04-proposer-checks-before-filing.md) |
| [x]  | 6 | [Research rechecks and refreshes](./issues/06-research-rechecks-premise.md) | [#4](./issues/04-proposer-checks-before-filing.md) |
| [x]  | 7 | [The build refuses a stale premise](./issues/07-build-refuses-stale-premise.md) | [#4](./issues/04-proposer-checks-before-filing.md) |
| [x]  | 8 | [Older proposals are not exempt](./issues/08-legacy-proposals-checked.md) | [#4](./issues/04-proposer-checks-before-filing.md) |

## User story coverage

| Slice | PRD stories |
|---|---|
| 1 | 20, 21, 30 |
| 2 | 19, 24 |
| 3 | 1, 2, 26 |
| 4 | 3, 4, 15, 22, 23 |
| 5 | 12, 13, 14, 34 |
| 6 | 5, 7, 8, 9, 10, 11, 25, 26, 27, 29 |
| 7 | 6, 7, 8, 9, 10, 11, 28, 29 |
| 8 | 16, 17, 18 |

**Not covered by any slice: stories 31, 32, 33.** They describe the
first-commit attribution fix and the benchmark baseline that belong to
the separate issue #30 salvage, which the PRD's Out of Scope section
names as its own work. The gap is recorded here so it stays visible
rather than reading as covered.

## Notes

- Slices 5, 6, 7 and 8 all depend only on slice 4, so once the verifier
  exists they can proceed in parallel or in any order.
- No slice requires a human handoff. Every slice is verifiable by the
  session that builds it, using `make test` or a dry run of the program
  it changes.
- Do not commit or push. That stays with the operator.
