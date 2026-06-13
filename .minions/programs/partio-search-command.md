---
id: partio-search-command
target_repos:
  - cli
acceptance_criteria:
  - partio search accepts a query string and searches checkpoint transcript content
  - Results display matching checkpoint IDs, associated commit SHAs, and relevant transcript excerpts
  - Search works against local checkpoint data on the orphan branch without requiring a remote
  - Results are sorted by relevance or recency
  - --json flag outputs machine-readable results
pr_labels:
  - minion
---

# Add `partio search` command for checkpoint content search

## Summary

Add a `partio search` command that performs full-text search across checkpoint transcript content stored on the orphan branch. Users should be able to find past sessions by searching for prompts, code snippets, or reasoning captured in checkpoint transcripts.

## Why

As checkpoint history grows, users need a way to find specific past sessions without manually browsing. A developer debugging a regression may want to search "why did we change the auth middleware" to find the session that introduced the change. This is the retrieval counterpart to Partio's capture workflow.

## Context

- Checkpoint data is stored as blobs on `partio/checkpoints/v1` orphan branch
- Session transcripts are JSONL files within checkpoint trees
- The upstream project (entireio/cli) added `entire search` in v0.5.4 and iteratively improved its TUI through v0.6.0
- Partio's initial implementation can be simpler — plain text search with formatted output, no TUI required initially
- Search should read checkpoint tree contents via git plumbing (`ls-tree`, `cat-file`) to avoid checkout

## References

- entireio/cli changelog 0.5.4: "`entire search` command now available with improved TUI usability"
- entireio/cli changelog 0.6.0: "`entire search` TUI gains unified palette with activity view, markdown rendering, and completions"
- entireio/cli#1171: Search bug report showing the feature's maturity level
