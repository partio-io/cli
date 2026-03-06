# Plan: GitHub Repo Polish for Partio Launch

## Context

Partio launches this week (March 3-7, 2026). The GitHub repos are the "storefront" — developers will decide in 10 seconds whether to star/try the tool. The CLI README is functional but lacks visual polish (no badges, no demo, missing the `resume` command, confusing "entire.io" tagline). The app README needs minor improvements. Neither repo has a CONTRIBUTING.md.

## What We'll Do

### 1. Rewrite CLI README (`cli/README.md`)

**Current state:** 121 lines, functional but plain. Missing badges, demo, `resume` command in table, has confusing "partial version of entire.io" tagline.

**Changes:**

- **Remove** the "partial version of entire.io" line — meaningless to OSS visitors
- **Add badges** after title: MIT License, Go Report Card, Latest Release
- **Rewrite intro** to two short paragraphs: what it does + privacy statement ("Nothing leaves your machine")
- **Add demo GIF placeholder** (`assets/demo.gif`) — we'll create a VHS tape file to generate it
- **Reorder Install** — Homebrew first (easier), Go second, add "Requires Go 1.25+" note
- **Simplify Quick Start** to exactly 5 steps: enable → commit → status → rewind --list → resume
- **Add `resume` to commands table** (with `--print` and `--copy` flags — confirmed from `resume.go`)
- **Tighten "How It Works"** — minor wording fixes (clarify it's the post-commit hook that creates checkpoints)
- **Collapse** Checkpoint Data, Configuration, and Git Worktrees into `<details>` blocks to keep README scannable
- **Add Contributing** section linking to `.github/CONTRIBUTING.md`

**Final section order:**
1. Title + badges
2. One-line tagline + description paragraph + privacy line
3. Demo GIF
4. Install (Homebrew, then Go)
5. Quick Start (5 steps)
6. Commands table (12 rows including resume variants)
7. How It Works (6 steps)
8. Collapsed: Checkpoint Data Structure
9. Collapsed: Configuration
10. Collapsed: Git Worktrees
11. Contributing link
12. License

### 2. Create CLI CONTRIBUTING.md (`cli/.github/CONTRIBUTING.md`)

Short, welcoming guide covering:
- Development setup (`git clone` + `make build`)
- Running tests (`make test`)
- Project structure (one-liner per directory)
- Making changes (fork → branch → test → lint → PR)
- Code style (one concern per file, stdlib testing only, minimal deps)
- Reporting issues (include `partio doctor` output)
- License agreement

### 3. Create VHS tape file (`cli/demo.tape`)

A declarative script for [charmbracelet/vhs](https://github.com/charmbracelet/vhs) that records a terminal GIF showing:
- `partio enable` in a repo
- Making a commit (simulated)
- `partio status` showing a captured checkpoint
- `partio rewind --list` showing checkpoint history
- `partio resume <id> --print` showing the composed context

**Note:** Generating the actual GIF requires VHS installed and a repo with real checkpoint data. We'll create the tape file and `assets/` directory; you'll generate the GIF manually before pushing.

### 4. Polish App README (`app/README.md`)

- **Add MIT badge** after title
- **Improve opening line** to cross-link to CLI repo: "A dashboard for browsing AI agent checkpoints captured by [partio](https://github.com/partio-io/cli)."
- **Add screenshot placeholder** (`assets/screenshot.png`)

### 5. Create App CONTRIBUTING.md (`app/.github/CONTRIBUTING.md`)

Short guide covering: clone → `.env.local` setup → `npm install` → `npm run dev`. Mention code conventions (App Router, `"use client"`, SWR hooks, `auth()` in API routes).

### 6. Create GitHub Release (v0.1.0) — instructions only

This requires pushing a tag to trigger the existing GoReleaser workflow. We'll:
- Prepare a release notes file (`cli/release-notes-v0.1.0.md`) for use with `gh release edit`
- Document the exact steps: merge to main → tag → push → edit release notes

**Cannot be done programmatically in this session:** merging branches, pushing tags, editing releases on GitHub.

### 7. Items requiring GitHub UI (documented, not automated)

| Task | How |
|------|-----|
| Pin `cli` and `app` repos | github.com/partio-io → "Customize your pins" |
| Social preview images | Each repo → Settings → General → Social preview (1280x640px) |
| Verify HOMEBREW_TAP_TOKEN secret | CLI repo → Settings → Secrets → Actions |

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `cli/README.md` | Rewrite |
| `cli/.github/CONTRIBUTING.md` | Create |
| `cli/demo.tape` | Create |
| `cli/assets/.gitkeep` | Create (directory for demo.gif) |
| `app/README.md` | Edit (badge, description, screenshot placeholder) |
| `app/.github/CONTRIBUTING.md` | Create |

## Verification

1. Read the CLI README on GitHub after push — badges render, demo GIF placeholder shows, collapsible sections work
2. Click CONTRIBUTING.md links from README — they resolve correctly
3. `partio resume` appears in the commands table with correct flags
4. App README cross-links to CLI repo correctly
5. After tagging v0.1.0: `brew install partio-io/tap/partio` installs the new version, release page shows binaries and notes
