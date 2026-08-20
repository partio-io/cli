# Issues — 2026-08-16-first-commit-attribution

Source: [prd.md](./prd.md)

| Done | # | Title | Blocked by |
|------|---|-------|------------|
| [x]  | 1 | [First commit records its lines](./issues/01-first-commit-records-lines.md) | None |
| [x]  | 2 | [First checkpoint lists its files](./issues/02-first-checkpoint-lists-files.md) | [#1](./issues/01-first-commit-records-lines.md) |
| [x]  | 3 | [First checkpoint stores its diff](./issues/03-first-checkpoint-stores-diff.md) | [#1](./issues/01-first-commit-records-lines.md) |
| [x]  | 4 | [A failed measurement is visible](./issues/04-failed-measurement-is-visible.md) | [#1](./issues/01-first-commit-records-lines.md) |
| [x]  | 5 | [The speed claim is measurable](./issues/05-speed-claim-is-measurable.md) | [#1](./issues/01-first-commit-records-lines.md) |

Slice 5 — pull request #653: decision pending. The operator records the
outcome here.

## User story coverage

| Slice | PRD stories |
|---|---|
| 1 | 1, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 16, 17, 29, 30, 31, 32, 33, 34 |
| 2 | 2, 4, 17, 29, 30, 32, 33, 34 |
| 3 | 3, 4, 17, 29, 30, 32, 33, 34 |
| 4 | 14, 15, 29, 30, 32 |
| 5 | 18, 19, 20, 21, 22, 23, 24, 28, 30 |

**Not covered by any slice: stories 25, 26 and 27.**

- Stories 25 and 26 ask that the false premise of issue #30 and the
  merge regression in pull request #653 are recorded. The PRD's Further
  Notes section already records both, so no slice adds them.
- Story 27 asks that the premise-gate document is corrected. The PRD
  names that as a separate decision in Out of Scope. It is unbuilt, and
  it is recorded here so the gap stays visible.

## Notes

- Slice 1 carries the shared constant and the shared helper. Slices 2,
  3, 4 and 5 all depend only on slice 1, so once it lands they can
  proceed in parallel or in any order.
- Slice 5 has a `## Handoff`. It is the only slice that ends with an
  action for the operator.
- Pull request #653 stays open. No slice closes it, comments on it, or
  pushes to its branch. That decision belongs to the operator.
- Do not commit or push. That stays with the operator.
