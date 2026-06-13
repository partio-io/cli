---
id: actionable-hook-error-messages
target_repos:
  - cli
acceptance_criteria:
  - Common hook failure modes are classified into distinct error types with user-facing messages
  - Each error type includes a short description and a suggested remediation action
  - Hook error output includes the error classification when running with default log level
  - At least these failure modes are classified: session not found, JSONL parse failure, git plumbing error, permission denied, checkpoint branch missing
  - Unit tests verify that each classified error produces the expected message and suggestion
pr_labels:
  - minion
---

# Classify common hook and checkpoint errors into actionable user messages

## Summary

Replace generic error logging in Partio's hook implementations with classified, actionable error messages that tell users what went wrong and how to fix it.

## Motivation

When Partio hooks fail during git operations, users currently see raw error strings from Go's error wrapping (e.g., "failed to create checkpoint: exit status 128") with no guidance on what to do. Similar tools (entireio/cli#963) found that classifying CLI errors into categories with remediation suggestions dramatically reduces support burden and user frustration.

Partio's hooks run silently by design (to avoid blocking git), but when something does go wrong, the logged warning should be immediately actionable — not require users to search documentation or open issues.

## Implementation hints

- Define an error classification type in `internal/hooks/` or a shared errors package with fields: `Code`, `Message`, `Suggestion`
- Classify at least: `SessionNotFound` ("No active agent session detected" / "Run `partio status` to check agent detection"), `JSONLParseError` ("Failed to parse session transcript" / "Check if the agent session file is corrupted"), `GitPlumbingError` ("Git object write failed" / "Run `partio doctor` to check repository health"), `PermissionDenied` ("Cannot write to git object database" / "Check file permissions on .git/objects/"), `CheckpointBranchMissing` ("Checkpoint branch not found" / "Run `partio enable` to reinitialize")
- Wrap existing error returns in hook implementations with the classifier
- Surface the classification in log output at warn level (visible by default)
- Keep the raw error available at debug level for advanced troubleshooting
