---
id: local-checkpoint-search-fallback
target_repos:
  - cli
acceptance_criteria:
  - partio search <query> searches checkpoint content on the local orphan branch when no remote API is configured
  - Search uses git log and git show on partio/checkpoints/v1 to grep checkpoint transcript and metadata blobs
  - Results include commit hash, checkpoint timestamp, and matching context snippet
  - Search returns results for prompt text, transcript content, and file paths stored in checkpoints
  - Flag --local forces local search even when a remote is configured
pr_labels:
  - minion
---

# Add local-first checkpoint search fallback

## Summary

Add a `partio search` command (or `--local` flag) that searches checkpoint content by grepping the local `partio/checkpoints/v1` orphan branch directly, without requiring a remote search API.

## Context

Currently Partio stores rich checkpoint data (prompts, transcripts, file lists, attribution metadata) on the local orphan branch, but provides no way to search through it. Users who want to find "which session modified file X" or "which checkpoint discussed topic Y" have no CLI-native way to do so.

Inspired by entireio/cli#1210 which added a `--local` fallback that greps the local checkpoint branch when no remote search API is available.

## Approach

- Walk the commit history of `partio/checkpoints/v1` using git log
- For each checkpoint commit, read the tree and search blob contents (transcript JSONL, metadata JSON) for the query string
- Return matching checkpoints with context: commit hash, timestamp, matched line/snippet
- Support basic substring and regex matching
- Add a `--local` flag to explicitly use local search
- Keep it fast by using git plumbing (cat-file, ls-tree) consistent with existing checkpoint storage patterns
