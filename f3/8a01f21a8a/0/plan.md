# Enable entire.io in CLI and Docs

## Context

The project has four sub-projects: **app**, **site**, **cli**, and **docs**. Entire.io is already enabled in app and site (both have `.entire/` config dirs and `.claude/settings.json` with entire hooks). The CLI and docs are missing both configurations. The goal is to add entire.io support to the CLI and docs to match the pattern already established in app and site.

## Current State

| Component | `.entire/` config | `.claude/settings.json` (hooks) |
|-----------|-------------------|---------------------------------|
| **app**   | Yes               | Yes                             |
| **site**  | Yes               | Yes                             |
| **cli**   | No                | No                              |
| **docs**  | No                | No                              |

## Changes

### 1. Create `cli/.entire/settings.json`

Copy the same pattern from `app/.entire/settings.json`:
```json
{
  "strategy": "manual-commit",
  "enabled": true,
  "telemetry": false
}
```

### 2. Create `cli/.entire/.gitignore`

Same as `app/.entire/.gitignore`:
```
tmp/
settings.local.json
metadata/
logs/
```

### 3. Create `cli/.claude/settings.json`

Same hooks pattern as `app/.claude/settings.json` — entire hooks for session tracking (post-task, post-todo, pre-task, session-end, session-start, stop, user-prompt-submit) plus deny permission on `.entire/metadata/`.

### 4. Create `docs/.entire/settings.json`

Same as above.

### 5. Create `docs/.entire/.gitignore`

Same as above.

### 6. Create `docs/.claude/settings.json`

Same hooks pattern as above.

## Files to Create (6 new files)

- `cli/.entire/settings.json`
- `cli/.entire/.gitignore`
- `cli/.claude/settings.json`
- `docs/.entire/settings.json`
- `docs/.entire/.gitignore`
- `docs/.claude/settings.json`

## Template Files (to copy from)

- `app/.entire/settings.json` — entire config
- `app/.entire/.gitignore` — gitignore pattern
- `app/.claude/settings.json` — Claude Code hooks

## Verification

- Confirm all 6 files are created with correct content
- Verify the entire hooks pattern matches across all 4 sub-projects (app, site, cli, docs)
