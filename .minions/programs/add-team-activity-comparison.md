---
id: add-team-activity-comparison
target_repos:
  - cli
acceptance_criteria:
  - "partio status or a new subcommand shows personal agent activity alongside team/repo-wide aggregates"
  - "Team data is derived from checkpoint branch metadata (commit authors, session counts) without external services"
  - "Output includes visual comparison (e.g., bar chart or percentage) of personal vs team checkpoint frequency"
  - "Works in single-user repos by showing only personal stats without errors"
  - "Table-driven tests cover both single-user and multi-contributor scenarios"
pr_labels:
  - minion
---

# Add team activity comparison to checkpoint status output

## Description

Add a team activity comparison view that shows the current user's agent-assisted commit activity alongside aggregated team statistics. This helps developers understand their AI agent usage patterns relative to the rest of the team.

The implementation should derive team data from the existing checkpoint branch (`partio/checkpoints/v1`) by scanning commit metadata (authors, session counts, checkpoint frequency) without requiring any external service or database.

## Why

As AI-assisted development becomes more common across teams, individual developers lack visibility into how their usage compares to teammates. This feature helps teams understand adoption patterns, identify opportunities for knowledge sharing, and track the overall impact of AI tooling on their codebase.

## Source

Inspired by entireio/cli PR #1015 — `entire recap` visual redesign with me/team comparison bars showing personal vs team activity metrics.

## Context hints

- `cmd/partio/` (CLI command definitions)
- `internal/checkpoint/` (checkpoint storage and querying)
