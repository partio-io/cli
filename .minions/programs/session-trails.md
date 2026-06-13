---
id: session-trails
target_repos:
  - cli
acceptance_criteria:
  - Users can create a named trail with `partio trail create <name>`
  - Checkpoints created while a trail is active are tagged with the trail name in metadata
  - `partio trail list` shows all trails with session/commit counts
  - `partio trail show <name>` displays the chronological sequence of sessions and commits in a trail
  - Trail metadata is stored on the checkpoint orphan branch alongside existing checkpoint data
  - Trails work correctly across branches (e.g., feature branch sessions and main branch sessions in the same trail)
pr_labels:
  - minion
---

# Add session trails for grouping related sessions across commits

## Problem

Partio captures individual sessions and links them to commits, but there's no way to group related sessions that contribute to the same feature or task. When a developer works on a feature across multiple commits, branches, and agent sessions over days or weeks, the full story is fragmented across individual checkpoints.

## Solution

Add a "trail" concept that groups related sessions into a named thread. A trail is a lightweight named entity stored in checkpoint metadata that links multiple sessions and commits together under a common thread.

### Key behaviors

- `partio trail create <name>` creates a new trail and marks it as active in `.partio/state/`
- `partio trail stop` deactivates the current trail
- While a trail is active, all new checkpoints include a `trail: <name>` field in their metadata
- `partio trail list` shows all trails with summary info (session count, date range, commits)
- `partio trail show <name>` displays the chronological sequence of sessions and commits in the trail
- Trail state is local (stored in `.partio/state/`) but trail metadata travels with checkpoints on the orphan branch

### Implementation hints

- Store active trail name in `.partio/state/active-trail.json`
- Add `trail` field to checkpoint metadata in `internal/checkpoint/`
- Post-commit hook reads active trail from state and includes it in checkpoint metadata
- Trail listing is a scan of checkpoint metadata on the orphan branch, filtered by trail field
- Use git plumbing (consistent with existing checkpoint storage) to read trail data

## Context

Inspired by entireio/cli's trail concept (PR #919 `entire trail link`, changelog 0.6.1 `entire labs review` with session support). Adapted for Partio's git-plumbing-based architecture.
