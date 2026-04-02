---
id: implement
target_repos:
  - cli
acceptance_criteria:
  - "All changes match what the issue describes"
  - "make test passes"
  - "make lint passes"
  - "PR description clearly explains what changed and why"
  - "Commit messages are meaningful and describe the change"
pr_labels:
  - minion
---

# Implement the issue

Read the issue provided as context and implement exactly what it describes.

Follow existing code patterns and conventions. Read the relevant code before making changes. Keep changes minimal — implement what the issue asks for, nothing more.

After implementation, run the appropriate checks (`make test`, `make lint`) and fix any failures.

## PR and commit quality

A human will review your PR. Make it easy for them:

- **Commit messages** should describe what changed and why, not just "implement feature". Use conventional commit format (e.g., `feat: add codex agent detection`).
- **PR title** should be a clear summary of what the PR does (not the issue title verbatim if it doesn't make sense as a PR title).
- **PR description** must include:
  - A summary of what was implemented
  - Key design decisions you made
  - How to test the changes
  - Reference to the source issue (e.g., "Resolves #120")

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
