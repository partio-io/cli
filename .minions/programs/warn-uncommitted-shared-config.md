---
id: warn-uncommitted-shared-config
target_repos:
  - cli
acceptance_criteria:
  - "After `partio enable` or `partio configure` modifies `.partio/settings.json`, a hint is printed reminding the user to commit the file if it contains team-shared settings"
  - "The hint only appears when `.partio/settings.json` has uncommitted changes (not when it is already tracked and clean)"
  - "The hint is suppressed in non-interactive contexts (e.g., CI) or when stdout is not a terminal"
  - "A test verifies the hint appears after a config modification when the file is untracked"
  - "A test verifies the hint does not appear when the file is already committed and unchanged"
pr_labels:
  - minion
  - feature
---

# Warn when shared config file has uncommitted changes

After `partio enable` or `partio configure` modifies `.partio/settings.json`, print a hint reminding the user to commit the file so that team members share the same configuration.

## Why

When users configure team-shared settings (like checkpoint strategy or agent type), forgetting to commit `.partio/settings.json` means teammates won't pick up those settings. This is especially easy to miss because the file is created/modified by CLI commands rather than edited by hand. A simple hint at the right moment prevents configuration drift across the team.

## Approach

- After any command that writes to `.partio/settings.json`, check if the file has uncommitted changes using `git status --porcelain`
- If the file is untracked or modified, print a hint: `Hint: .partio/settings.json has uncommitted changes. Consider committing it so your team shares the same config.`
- Only show the hint when stdout is a terminal (skip in CI/piped contexts)
- Keep the hint non-blocking — it's informational, not an error
