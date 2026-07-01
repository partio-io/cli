---
id: clipboard-device-code-login
target_repos:
  - cli
acceptance_criteria:
  - When device code auth flow starts, the code is automatically copied to the system clipboard
  - A message informs the user the code was copied
  - If clipboard access fails (e.g., headless), the flow continues normally with manual copy instructions
  - Works on macOS (pbcopy), Linux (xclip/xsel), and Windows (clip.exe)
pr_labels:
  - minion
  - enhancement
---

# Copy device code to clipboard during auth login

When `partio` uses a device-code authentication flow (e.g., for checkpoint remote auth), users must manually copy a code and paste it into a browser. This is a minor friction point that can be eliminated by auto-copying to clipboard.

## What to implement

1. After generating/receiving the device code, attempt to copy it to the system clipboard
2. Use platform-appropriate commands: `pbcopy` (macOS), `xclip -selection clipboard` (Linux), `clip.exe` (Windows)
3. Print a message like: `Device code copied to clipboard. Open <URL> and paste to authenticate.`
4. If clipboard copy fails, fall back to the existing behavior (print code, ask user to copy manually)

## Why this matters

Small UX improvements in auth flows reduce friction during initial setup — the moment when users form their first impression of the tool.

## Source

Inspired by entireio/cli PR #1093 — copy login device code to clipboard.
