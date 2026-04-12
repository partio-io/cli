---
id: checkpoint-v2-migration-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio migrate --checkpoints v2` converts existing v1 checkpoint data to v2 format"
  - "Migration reads all commits from `partio/checkpoints/v1` and writes compact `transcript.jsonl` + metadata to the v2 ref layout"
  - "Migration is idempotent — running it twice does not duplicate or corrupt data"
  - "Progress output shows how many checkpoints were migrated"
  - "Dry-run flag (`--dry-run`) previews what would be migrated without writing"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add `partio migrate` command for checkpoint format upgrades

Add a `partio migrate --checkpoints v2` CLI command that migrates existing checkpoint data from the v1 orphan branch format to the v2 compact format.

## Context

Partio currently stores checkpoints on `partio/checkpoints/v1` using git plumbing commands. A v2 format with compact `transcript.jsonl` and metadata on a `/main` ref layout would reduce storage overhead and improve query performance. Before v2 can become the default, users need a migration path for existing checkpoint data.

## What to implement

1. A new `partio migrate` Cobra command under `cmd/partio/`
2. A `--checkpoints` flag accepting `v2` as the target format
3. A `--dry-run` flag that shows what would be migrated without writing
4. Migration logic that:
   - Reads checkpoint commits from the v1 branch
   - Extracts session data and transcript content
   - Writes compact transcript.jsonl files and metadata to the v2 ref layout
   - Preserves all existing checkpoint metadata (session IDs, attribution, timestamps)
5. Progress output showing migration status

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
