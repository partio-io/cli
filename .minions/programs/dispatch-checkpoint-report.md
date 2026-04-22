---
id: dispatch-checkpoint-report
target_repos:
  - cli
acceptance_criteria:
  - "`partio dispatch` command generates a markdown summary of recent checkpoint activity"
  - "Supports `--since` flag to control the time window (default: 24h)"
  - "Supports `--branch` flag to filter by branch (default: current branch)"
  - "Output includes commit hashes, agent attribution, session prompts, and checkpoint counts"
  - "Works with local checkpoint data only (no API dependency)"
  - "Respects existing session/checkpoint storage format on the orphan branch"
pr_labels:
  - minion
---

# Add `partio dispatch` command to generate checkpoint activity reports

## Description

Add a `partio dispatch` command that reads recent checkpoints from the orphan branch (`partio/checkpoints/v1`) and generates a structured markdown summary of agent activity. This gives developers a quick overview of what AI agents have done in a repository over a given time window — useful for standup summaries, PR descriptions, and team visibility.

The command should:
- Walk checkpoint commits on the orphan branch within the specified time window
- Extract metadata (agent name, session prompt, files touched, attribution percentages)
- Group activity by session and render a readable markdown report to stdout
- Support `--since` (duration, e.g. `24h`, `7d`) and `--branch` (git branch name) flags

## Why

Partio captures rich checkpoint data but currently has no way to generate a human-readable summary of recent agent activity. Developers working with AI agents need a quick way to see what happened — for standups, PR context, or audit trails. This complements `partio status` (which shows current state) with a historical view.

## Source

Inspired by entireio/cli PR #1004 (`entire dispatch`) which adds a similar report generation command.

## Context hints

- `internal/checkpoint/` — checkpoint domain types and git plumbing storage
- `cmd/partio/` — CLI command registration
- `internal/session/` — session metadata
