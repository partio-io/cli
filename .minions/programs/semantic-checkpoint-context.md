---
id: semantic-checkpoint-context
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes entity-level changes (functions, types, methods added/removed/modified) when available
  - Entity detection works for Go source files at minimum
  - Existing checkpoint consumers are unaffected (semantic data is additive)
  - partio rewind output shows entity-level change summary when semantic data is present
  - Graceful fallback to file-level output when semantic parsing fails or is unavailable
pr_labels:
  - minion
---

# Add semantic entity-level context to checkpoint metadata

## Summary

Enhance checkpoint metadata with semantic code analysis that identifies which code entities (functions, types, methods) were added, removed, or modified in each checkpoint — going beyond file-level diffs to show structural changes.

## Motivation

Partio captures the *why* behind code changes, but checkpoint metadata currently only tracks file-level diffs. When reviewing checkpoint history, users see "main.go changed" but not "function ParseSession was added and method Detect had its signature changed." Entity-level context makes checkpoint history significantly more useful for understanding what actually happened.

## Implementation Notes

- Parse diffs using Go's `go/ast` and `go/parser` (no external dependency needed for Go files) to extract entity-level changes
- Store semantic changes as an optional field in checkpoint metadata (additive, non-breaking)
- Integrate into `partio rewind` output to show concise entity summaries alongside file-level changes
- Start with Go support since the CLI is Go-only; design the interface to be extensible for other languages later
- Keep semantic analysis deterministic and local — no LLM or network calls

## Source

Inspired by entireio/cli#1283 which adds tree-sitter based semantic diffs to checkpoint explain/rewind.
