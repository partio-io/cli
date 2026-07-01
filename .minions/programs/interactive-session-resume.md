---
id: interactive-session-resume
target_repos:
  - cli
acceptance_criteria:
  - "partio resume with no arguments shows a selectable list of stopped/idle sessions"
  - "The picker displays session ID, agent name, start time, and last activity"
  - "Selecting a session resumes it (existing resume behavior)"
  - "If only one session exists, it is selected automatically without showing the picker"
  - "If no sessions exist, a clear message is shown"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add interactive session resume picker

## Summary

When `partio resume` is run without specifying a session ID, show an interactive picker listing all stopped or idle sessions so the user can select which one to resume.

## Background

Inspired by entireio/cli changelog 0.7.7 which added an "interactive resume picker for stopped/idle sessions." Currently `partio resume` requires the user to know the session ID. An interactive picker makes it easier to find and resume the right session.

## Implementation notes

- Extend the existing `cmd/partio/resume.go` command to detect when no session ID argument is provided
- Query available sessions from the session manager (sessions in stopped/idle state)
- Display a simple list with session metadata (ID, agent, timestamps) and let the user pick one
- If only one session is available, auto-select it
- Keep the existing behavior when a session ID is explicitly provided
