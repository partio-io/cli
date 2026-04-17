---
id: compact-checkpoint-trailer
target_repos:
  - cli
acceptance_criteria:
  - "Partio-Checkpoint trailer value uses a short hash (first 12 characters) instead of a full reference"
  - "Existing checkpoint lookup/rewind commands can resolve both short and full trailer values"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Use compact checkpoint trailer format to reduce noise in git UIs

## Problem

The `Partio-Checkpoint` trailer appended to commit messages can be visually noisy in GitHub's commit UI and other git tools. Long reference values clutter the commit view, especially when browsing commit history.

## Solution

Shorten the `Partio-Checkpoint` trailer value to use a 12-character short hash instead of the full reference. Ensure that checkpoint lookup commands (`partio rewind`, etc.) can resolve both short and full-length trailer values for backwards compatibility.

## Context

- Inspired by entireio/cli#868 (community feedback about trailer noise)
- Relevant code: `internal/checkpoint/` (trailer writing), `cmd/partio/` (rewind command)

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
