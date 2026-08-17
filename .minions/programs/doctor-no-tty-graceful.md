---
id: doctor-no-tty-graceful
target_repos:
  - cli
acceptance_criteria:
  - Running `partio doctor` without a TTY (piped output, CI, or from inside a Claude Code session) does not crash or hang
  - In non-TTY mode, any interactive selection prompt (e.g., session repair choices) is replaced with a plain-text report listing actionable items
  - `partio doctor` exits with a non-zero code if actionable problems are found, in both TTY and non-TTY modes
  - `partio doctor --force` skips all interactive confirmation prompts regardless of TTY
  - A unit test exercises the non-TTY path and asserts no panic and correct exit code
pr_labels:
  - minion
---

# `partio doctor` should not crash or hang when run without a TTY

`partio doctor` is a natural candidate for calls from inside Claude Code sessions (the agent
diagnosing its own environment) and from CI. If any part of the doctor flow uses interactive
components that require a terminal — selection prompts, pager output, Bubble Tea TUI — it will
crash with an unhelpful error like `open /dev/tty: device not configured` or hang waiting for
input that never arrives.

Guard every interactive prompt behind an `isatty` check. In non-TTY mode:
- Print a plain-text diagnostic report with the same checks the TUI would show
- Skip any repair prompts; describe what would be fixed and the manual command to run it
- Return exit code 0 if everything is healthy, non-zero if actionable issues are found

Add `--force` to skip interactive confirmation for repair actions even in TTY mode (so
automation can invoke doctor with `--force` to auto-repair without interactive approval).

Inspired by entireio/cli PR #2015 "fix(doctor): hint --force instead of crashing on stuck
sessions with no TTY", which hit this bug live when doctor was called from inside a Claude Code
session on 2026-08-17.

## Context hints

- `cmd/partio/` — doctor command implementation
- `internal/session/` — session repair logic that may use interactive prompts

<!-- program: .minions/programs/doctor-no-tty-graceful.md -->
