---
id: track-attach-reason-on-checkpoints
target_repos:
  - cli
acceptance_criteria:
  - Define an AttachReason type with named constants for each checkpoint creation path (e.g. active_session, manual_commit, file_overlap)
  - Post-commit hook records the attach reason when creating a checkpoint
  - AttachReason is persisted in checkpoint metadata JSON
  - partio status or checkpoint inspection shows the attach reason
  - Table-driven tests verify each attach reason is correctly assigned and persisted
pr_labels:
  - minion
---

# Track attach reason on checkpoint metadata

## Summary

Record **why** each checkpoint was created by persisting an `attach_reason` field in checkpoint metadata. Currently, Partio's post-commit hook creates checkpoints when an agent session is detected, but there is no structured record of which decision path led to the checkpoint being created. This makes it difficult to debug checkpoint behavior or understand checkpoint provenance after the fact.

## Motivation

When debugging checkpoint issues or auditing checkpoint history, the only signal is binary: a checkpoint exists or it doesn't. Users and maintainers cannot tell whether a checkpoint was created because:
- The agent had an active session with recent interaction
- Files in the commit overlapped with the agent's touched files
- The user explicitly triggered checkpoint creation
- Some other heuristic matched

Adding a typed reason field makes checkpoint creation transparent and debuggable.

## Implementation Notes

- Add an `AttachReason` string type in `internal/checkpoint/` with named constants
- Extend the checkpoint metadata structure to include `attach_reason`
- Set the reason in the post-commit hook based on which detection path succeeded
- Include a single structured debug log line with the attach decision and reason
- Persist the reason when writing checkpoint metadata via git plumbing

## Source

Inspired by entireio/cli#1199 which adds `session.AttachReason` with typed constants and persists it on `CommittedMetadata.AttachReason` in `metadata.json`.
