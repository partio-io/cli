# Issues — 2026-08-03-minion-pr-audits

Source: [prd.md](./prd.md)

| Done | # | Title | Blocked by |
|------|---|-------|------------|
| [x]  | 1 | [Lint gate for swallowed errors](./issues/01-lint-gate.md) | None |
| [x]  | 2 | [Dead-code audit, read-only](./issues/02-deadcode-audit-readonly.md) | None |
| [x]  | 3 | [Dead-code fix round](./issues/03-deadcode-fix-round.md) | [#2](./issues/02-deadcode-audit-readonly.md) |
| [x]  | 4 | [E2e-need audit](./issues/04-e2e-need-audit.md) | [#2](./issues/02-deadcode-audit-readonly.md) |
| [x]  | 5 | [Session-liveness proposal](./issues/05-session-liveness-proposal.md) | None |

Shared names all slices must agree on (defined in #2, reused everywhere):

- Skip label: `no-audit`
- Audit comment body prefix: `Minion audit —`
- Audit fix-commit subject prefix: `audit:`
- Verdict file: `.minion-audit/verdict.json` in the workflow workspace
  (never committed)
