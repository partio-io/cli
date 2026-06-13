---
id: hunk-level-attribution
target_repos:
  - cli
acceptance_criteria:
  - "Calculate returns per-file attribution based on git diff hunks and agent session timing"
  - "Result struct includes per-file breakdown with agent vs human line counts"
  - "Mixed commits (some hunks authored while agent active, some not) get proportional attribution"
  - "Binary attribution (0%/100%) remains the fallback when session timing data is unavailable"
  - "make test passes with table-driven tests covering mixed, agent-only, and human-only scenarios"
  - "make lint passes"
pr_labels:
  - minion
---

# Add hunk-level attribution for mixed agent/human commits

## Problem

Partio currently uses binary attribution: if an agent was detected during a commit, 100% of lines are attributed to the agent; otherwise 0%. This is inaccurate for commits that contain a mix of agent-authored and human-authored changes — for example, when a developer makes manual edits after an agent session within the same commit.

## What to implement

Enhance the attribution calculation in `internal/attribution/` to analyze individual diff hunks and correlate them with agent session activity windows:

1. **Parse git diff hunks** — Break the commit diff into per-file, per-hunk segments using `git diff` output
2. **Correlate with session timing** — When session data includes timestamps, compare hunk authorship windows against agent session active periods
3. **Per-file breakdown** — Add a `Files []FileAttribution` field to `Result` with per-file agent/human line counts
4. **Graceful fallback** — When session timing data is unavailable, fall back to the current binary attribution behavior

## Context

- `internal/attribution/calculate.go` — Current binary calculation
- `internal/attribution/attribution.go` — Result type definition
- `internal/agent/claude/parse_jsonl.go` — Session data with timestamps
- Inspired by entireio/cli's work on clarifying human-added attribution (PR #1186) and the general trend toward more granular code provenance
