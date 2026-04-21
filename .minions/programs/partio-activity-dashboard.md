---
id: partio-activity-dashboard
target_repos:
  - cli
acceptance_criteria:
  - "`partio activity` reads checkpoint metadata from the `partio/checkpoints/v1` orphan branch and computes aggregate stats (total checkpoints, checkpoints per day, active streak length)"
  - "Renders an interactive Bubble Tea TUI with stat cards, a contribution scatter chart (commits by day), and a scrollable checkpoint list"
  - "Falls back to static text output when stdout is not a TTY (piped output)"
  - "All computation and rendering logic is covered by table-driven tests"
  - "No new external dependencies beyond `charmbracelet/bubbletea` and `charmbracelet/lipgloss`"
pr_labels:
  - minion
---

# Add `partio activity` command for interactive checkpoint activity dashboard

## Description

Add a `partio activity` command that provides an interactive terminal dashboard showing checkpoint activity for the current repository. Unlike `partio status` (which shows the current session state) or the proposed `partio insights` (#128, which prints text-based aggregate stats), this command provides a visual, interactive TUI for exploring checkpoint activity over time.

The dashboard reads checkpoint metadata directly from the `partio/checkpoints/v1` orphan branch using the existing git plumbing read path (same approach as `partio rewind` and `partio status`). No external API is needed since all data is local.

## Proposed layout

- **Stat cards row**: Total checkpoints, average per day, current streak (consecutive days with checkpoints), longest streak
- **Contribution chart**: ASCII scatter/heatmap showing checkpoint activity by day (similar to GitHub's contribution graph)
- **Recent checkpoints list**: Scrollable list showing commit hash, date, agent, and attribution percentage

## Implementation hints

- Use `charmbracelet/bubbletea` for the interactive TUI and `charmbracelet/lipgloss` for styling (these are the standard Go TUI libraries)
- Add `cmd/partio/activity.go` for the Cobra command
- Add `internal/checkpoint/activity.go` for reading and aggregating checkpoint metadata from the orphan branch
- Detect TTY with `os.Stdout.Fd()` and `term.IsTerminal()` to fall back to static output
- Follow the existing one-file-per-concern pattern
