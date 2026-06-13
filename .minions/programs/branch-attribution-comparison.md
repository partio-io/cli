---
id: branch-attribution-comparison
target_repos:
  - cli
acceptance_criteria:
  - New partio diff subcommand or --base flag on existing command compares attribution across a branch
  - Output shows aggregate agent vs human line counts for the branch relative to base
  - Per-file attribution breakdown is available in the output
  - Works with standard git branch topologies (feature branches off main)
  - JSON output mode available for machine consumption
pr_labels:
  - minion
---

# Add branch-level attribution comparison against base

## Summary

Add a command or flag that compares agent attribution across an entire branch against its merge base, showing aggregate and per-file agent vs human contribution for all commits on the branch.

## Motivation

Partio tracks per-commit attribution (agent vs human), but there's no way to see the aggregate picture across a feature branch. Teams reviewing PRs want to know "what percentage of this branch was agent-written?" without manually inspecting each checkpoint. This enables branch-level visibility into agent contribution, which is the natural unit for code review.

## Implementation Notes

- Use `git merge-base` to find the branch point, then aggregate attribution data from all checkpoints between the base and HEAD
- Show total lines added by agent vs human, with per-file breakdown
- Support `--base` flag to specify the comparison branch (default: main/master)
- Include `--json` output for CI/dashboard integration
- Build on existing attribution calculation in `internal/attribution/`

## Source

Inspired by entireio/cli 0.6.2 changelog: `entire review` default scope now compares against mainline with `--base` flag, providing branch-scoped analysis.
