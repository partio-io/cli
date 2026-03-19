# Plan: Add demo GIF to docs overview page

## Context
The demo GIF (`cli/assets/demo.gif`) is finalized. It's already referenced in the CLI README. Now we need to add it to the Mintlify docs site by creating a CLI overview page.

## Changes

### 1. Copy demo GIF to docs images
Copy `cli/assets/demo.gif` to `docs/images/cli/demo.gif` (following the existing pattern of `docs/images/app/` for app screenshots).

### 2. Create `docs/cli/overview.mdx`
New overview page with the demo GIF and a brief intro to the CLI. Keep it concise — the other CLI pages already cover installation, commands, configuration, and strategies in detail.

```mdx
---
title: Overview
description: "See partio in action"
---

# CLI Overview

partio hooks into Git to capture AI agent sessions alongside your commits. Here's what it looks like in practice:

<img src="/images/cli/demo.gif" alt="partio demo" />

When you commit, partio detects the active Claude Code session, captures the full conversation transcript, and stores it as a checkpoint on an orphan branch — all without leaving your terminal.

<CardGroup cols={2}>
  <Card title="Installation" icon="download" href="/cli/installation">
    Install via Homebrew or Go
  </Card>
  <Card title="Commands" icon="terminal" href="/cli/commands">
    Full CLI reference
  </Card>
</CardGroup>
```

### 3. Update `docs/mint.json` navigation
Add `cli/overview` as the first page in the "CLI Reference" group:

```json
{
  "group": "CLI Reference",
  "pages": [
    "cli/overview",
    "cli/installation",
    "cli/commands",
    "cli/configuration",
    "cli/strategies"
  ]
}
```

## Files to modify
- `docs/images/cli/demo.gif` (new — copy from `cli/assets/demo.gif`)
- `docs/cli/overview.mdx` (new)
- `docs/mint.json` (add `cli/overview` to navigation)

## Verification
- Run `cd docs && mintlify dev` and check the CLI Overview page renders with the GIF
- Verify navigation order in the sidebar
