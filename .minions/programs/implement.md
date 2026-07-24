---
id: implement
target_repos:
  - cli
slices: true
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

## Slice builds

An issue that carries a slice plan (a `minion:research-slices` comment) is built one slice per session. Your prompt then names your slice under "Your Slice" — build only that slice: implement its acceptance criteria and nothing from any other slice, even when adjacent code invites it. Earlier slices are already committed on your branch — build on their work, never redo or undo it. Later slices belong to their own sessions.

- The acceptance criteria above apply to your slice's changes, not the whole issue at once.
- Leave the tree green at your boundary: checks (`make test`, `make lint`) run after every slice.
- Commit your work with the same message quality as below; anything left uncommitted is swept into a generic `slice N/M` commit by the runtime.
- The runtime pushes each slice and opens the single PR after the final slice — never open a PR yourself.

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
