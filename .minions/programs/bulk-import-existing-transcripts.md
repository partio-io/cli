---
id: bulk-import-existing-transcripts
target_repos:
  - cli
pr_labels:
  - minion
acceptance_criteria:
  - A new partio import command discovers Claude Code JSONL transcripts under ~/.claude/projects/
  - The command matches transcripts to commits in the current repo by correlating timestamps and file paths
  - Each matched transcript is stored as a checkpoint on the orphan branch with appropriate metadata
  - Running import twice does not create duplicate checkpoints
  - The command reports how many transcripts were found, matched, and imported
  - Unit tests cover transcript discovery, commit matching, and deduplication logic
---

# Bulk import existing Claude Code transcripts into checkpoints

Add a `partio import` command that discovers pre-existing Claude Code transcripts from `~/.claude/projects/` and retroactively creates checkpoints for commits in the current repository.

## Context

Entireio/cli issue #1336 requests bulk import of existing Claude Code transcripts to improve cold start experience. Users who adopt Partio on an existing repo have months of Claude Code session history that was created before Partio's hooks were installed. Importing these transcripts would immediately provide the "why" context for historical commits.

## What to implement

1. Add a `partio import` command that scans `~/.claude/projects/` for JSONL transcript files.
2. For each transcript, extract timestamps and file modification events, then correlate with commits in the current repo's git log.
3. For matched transcripts, create checkpoints on the orphan branch using the existing checkpoint storage plumbing, with metadata indicating the checkpoint was retroactively imported.
4. Implement deduplication: skip transcripts that already have corresponding checkpoints (by matching commit SHA or session ID).
5. Support `--dry-run` flag to preview what would be imported without writing.
