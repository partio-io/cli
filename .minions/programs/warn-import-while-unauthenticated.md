---
id: warn-import-while-unauthenticated
target_repos: [cli]
acceptance_criteria:
  - If `partio enable` detects the user is not authenticated (no stored credentials) and proceeds to import existing sessions, it prints a prominent warning that the imported checkpoints are local-only and will not sync until the user authenticates
  - The warning includes the command the user should run after authenticating to push the locally-imported checkpoints
  - The enable flow does not abort — it completes the local import and surfaces the warning at the end
  - Unit test verifies the warning is emitted when the auth check returns unauthenticated and session import runs
pr_labels:
  - minion
---

# Warn when importing agent history while not authenticated

When `partio enable` runs for a repository and discovers existing agent session history to import, it should check whether the user is authenticated. If not, it should import the sessions into the local checkpoint store as normal, but prominently warn the user that the imported checkpoints are **local-only** and will not be visible to teammates or accessible on another device until they authenticate and push.

Without this warning, users who run `partio enable` before logging in end up with a local-only checkpoint store and no indication that anything is wrong — the import appears to succeed, but the data is silently stranded locally.

## Why

Silent local-only imports create a deceptive success state: `partio enable` exits 0, the checkpoints exist locally, but they will never sync. Users only discover the problem when they try to access history from a second device or share a repo and find the checkpoints are missing. A clear warning at enable time avoids the confusion and provides a concrete remediation path.

## User Relevance

New users often run `partio enable` as part of setting up a machine before completing authentication. This feature ensures they are not surprised later when their AI session history fails to appear on another device or in the web dashboard.

## Source

entireio/cli issue #1773 "entire enable/import while logged out imports locally with no way to sync after login" and CHANGELOG 0.9.0 ("Import gains imported sessions... and warns when importing agent history while logged out").

## Acceptance Criteria

- If `partio enable` detects the user is not authenticated (no stored credentials) and proceeds to import existing sessions, it prints a prominent warning that the imported checkpoints are local-only and will not sync until the user authenticates
- The warning includes the command the user should run after authenticating to push the locally-imported checkpoints
- The enable flow does not abort — it completes the local import and surfaces the warning at the end
- Unit test verifies the warning is emitted when the auth check returns unauthenticated and session import runs
