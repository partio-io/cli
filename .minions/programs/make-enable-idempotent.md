---
id: make-enable-idempotent
target_repos:
  - cli
acceptance_criteria:
  - "Running `partio enable` when `.partio/` already exists reinstalls hooks and ensures settings are consistent"
  - "Running `partio enable` after `partio disable` restores full functionality without manual workarounds"
  - "Hook scripts are regenerated if they are missing or stale, even when `.partio/settings.json` exists"
  - "The enabled field in settings.json is set to true after `partio enable` regardless of prior state"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Make `partio enable` idempotent

Make `partio enable` idempotent so it can be re-run safely to repair or update an existing installation without requiring `partio disable` first.

## Context

Currently `partio enable` skips hook installation if `.partio/` already exists (documented known limitation in CLAUDE.md). This means:
- If hooks are corrupted or removed by another tool, users must run `partio disable && partio enable` to fix them
- If settings get into an inconsistent state (e.g., `enabled: false` left behind), `partio enable` doesn't fix it

This is inspired by similar issues in entireio/cli (#1140, #1123) where the enable/disable lifecycle proved fragile in practice.

## Implementation guidance

- When `.partio/` exists, `partio enable` should still verify and reinstall hooks if needed
- Settings should be reconciled: ensure `enabled: true` is written regardless of prior state
- Existing valid configuration (strategy, agent, etc.) should be preserved — only repair what's broken
- Add tests covering: re-enable after disable, re-enable with missing hooks, re-enable with stale hooks

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
