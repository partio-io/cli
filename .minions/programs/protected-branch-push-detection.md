---
id: protected-branch-push-detection
target_repos:
  - cli
acceptance_criteria:
  - "Pre-push hook warns to stderr when pushing to a branch matching a protected branch pattern"
  - "Warning is only emitted when the tip commit has 100% agent attribution (Partio-Attribution trailer)"
  - "Default protected branch patterns include: main, master, develop, and release/*"
  - "'protected_branches' array in .partio/settings.json overrides defaults when present"
  - "Push is never blocked — only a warning is emitted and the hook exits 0"
  - "No warning is emitted when pushing to non-protected branches regardless of attribution"
pr_labels:
  - minion
---

# Detect and warn on pushes to protected branches

In the pre-push hook, detect when the user is pushing directly to commonly protected branch names (main, master, develop, release/*). When detected and the commit contains a Partio-Attribution trailer with 100% agent attribution, emit a prominent warning to stderr (but do not block the push).

The list of protected branch patterns should be configurable via `.partio/settings.json` under a `protected_branches` string array field with sensible defaults.

## Why

Pushing fully AI-attributed commits directly to protected branches without review is a risk. A warning at push time gives developers a chance to reconsider, without being overly restrictive.

## User relevance

Teams using branch protection policies benefit from an additional safety net reminding them when AI-written code is about to land on a protected branch without a PR review.

## Context hints

- `internal/hooks/prepush.go`
- `internal/config/config.go`
- `internal/config/defaults.go`
