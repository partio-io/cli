---
id: repo-rename-resilience
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint references remain valid and accessible after a GitHub repository rename
  - partio doctor detects and reports when the remote URL has changed since checkpoints were created
  - Checkpoint metadata stores enough information to survive remote URL changes (e.g., checkpoint branch ref names are repo-agnostic)
  - Documentation describes the rename workflow and any manual steps required
pr_labels:
  - minion
---

# Handle repository renames without breaking checkpoint links

## Summary

Ensure that Partio checkpoint data remains accessible and correctly linked after a GitHub repository is renamed. Currently, checkpoint metadata and refs may contain hardcoded repository references that break when the remote URL changes.

## Why

Repository renames are common during project evolution (prototype to production name, org transfers). Users who have been building checkpoint history should not lose access to that history after a rename. This is especially important for teams evaluating Partio — knowing their data survives organizational changes reduces adoption risk.

## Context

- Checkpoint refs (`partio/checkpoints/v1`) are stored locally and pushed to remotes
- Checkpoint metadata may reference the remote URL at creation time
- Git remotes update automatically for GitHub renames via redirects, but ref names and stored metadata do not
- The upstream project received a feature request about this: entireio/cli#909
- Partio's checkpoint storage uses git plumbing, so refs themselves are repo-agnostic — the main concern is any metadata fields that embed the repo name/URL

## References

- entireio/cli#909: "[FEATURE QUESTION] repository rename"
