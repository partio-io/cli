---
id: normalize-attribution-paths
target_repos:
  - cli
acceptance_criteria:
  - File paths from agent sessions are normalized to repo-relative format before comparison with git staged files
  - Absolute paths, paths with different separators, and paths with leading ./ are all correctly normalized
  - Attribution matching works correctly when the agent reports absolute paths but git reports relative paths
  - Tests cover cross-platform path variations (Unix absolute, Windows absolute, relative, with/without leading ./)
pr_labels:
  - minion
---

# Normalize file paths for cross-platform checkpoint attribution

## Summary

Normalize file paths reported by AI agents to repo-relative format before comparing with git staged files during attribution calculation. Agents may report absolute paths (e.g., `/Users/dev/project/src/main.go`) while git reports repo-relative paths (e.g., `src/main.go`), causing attribution matching to silently fail.

## Why

When path formats don't match, commits that should be attributed to an agent session get zero attribution — the hook runs, detects the session, but concludes "no content to link" because the file lists don't overlap. This is a silent data loss bug that undermines Partio's core value proposition: if attribution is wrong, the checkpoint history is misleading.

## Context

- Attribution is calculated in `internal/attribution/` by comparing agent-touched files with staged files
- Hook state includes files the agent touched, stored in `.partio/state/pre-commit.json`
- The upstream project fixed this in PR entireio/cli#779: "Normalize FilesTouched paths to prevent missing checkpoint trailers"
- Partio's current attribution is binary (0% or 100%), but path normalization is needed regardless — even binary attribution requires correct file overlap detection
- Different agents may report paths differently (absolute vs relative, OS-specific separators)

## References

- entireio/cli#779: "Fix: Normalize FilesTouched paths to prevent missing checkpoint trailers"
- entireio/cli#784: Related issue about only first commit getting trailers
