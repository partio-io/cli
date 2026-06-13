---
id: hosting-platform-branch-exclusion
target_repos:
  - cli
acceptance_criteria:
  - "partio enable detects Vercel (vercel.json), Netlify (netlify.toml), and similar hosting platform config files in the repo"
  - "When a hosting platform is detected, partio enable prompts the user to add partio/* branch ignore rules to the platform config"
  - "The ignore rules prevent hosting platforms from triggering deployments on checkpoint branches"
  - "partio doctor warns when a hosting platform is detected but branch exclusion is not configured"
  - "Non-interactive mode (CI) skips the prompt and emits a warning instead"
pr_labels:
  - minion
---

# Detect hosting platforms and exclude checkpoint branches from deployments

## Summary

When `partio enable` is run in a repository that deploys via Vercel, Netlify, or similar hosting platforms, checkpoint branches (`partio/checkpoints/v1`, `partio/*`) can trigger unwanted deployments. Partio should detect hosting platform configuration files and guide users to add branch exclusion rules.

## Motivation

Partio creates orphan branches for checkpoint storage. Hosting platforms like Vercel and Netlify monitor all branches by default and may attempt to build/deploy from checkpoint branches, causing:

- Wasted CI/CD minutes and build quota
- Failed deployments (checkpoint branches don't contain deployable code)
- Confusing deployment notifications for the team

This is a common pitfall for new users that can be prevented at setup time.

## Implementation Notes

- During `partio enable`, check for hosting platform config files (`vercel.json`, `netlify.toml`, `.vercel/`, etc.)
- If found, check whether checkpoint branch patterns are already excluded
- For Vercel: check `git.ignoredBranches` in `vercel.json` or project settings
- For Netlify: check `[build]` branch configuration in `netlify.toml`
- Prompt the user interactively to add the exclusion, or emit a warning in non-interactive mode
- Add a `partio doctor` check that warns about missing exclusion rules

## Source

Inspired by entireio/cli#904 — adds Vercel-specific branch exclusion detection to the enable flow.
