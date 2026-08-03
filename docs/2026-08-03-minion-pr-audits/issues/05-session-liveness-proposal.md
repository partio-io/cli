# 05 — Session-liveness proposal

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

Not an implementation — a proposal that routes the F1 fix through the
normal minion pipeline. Author the program file and file the
`minion-proposal` issue; the build happens later, when the operator
approves it.

The defect being proposed away: hook liveness asks "is any process
with 'claude' in its command line running anywhere on this machine?"
(a global process-name grep) instead of "is the agent that owns this
repo's session alive?". On a workstation where some agent is nearly
always running, session-end is effectively never recorded, the
condensed-session skip never engages, and a manual commit made after a
session ended can be rewritten with agent-attribution trailers and a
redundant checkpoint — false data in the product's core promise.

The proposed fix: the pre-commit hook already records the owning
agent's process ID in the session state. Session-end and skip
decisions should consult that PID's liveness instead of the global
grep. With the tautological check gone, the pre-commit
condensed-session block (left dead by PR #586's fix) becomes
meaningful again or is deleted outright — the proposal should name
that cleanup explicitly.

## User stories covered

23, 24.

## Acceptance criteria

- [ ] Program file exists under the programs directory with the
      pipeline's standard frontmatter (id, target repo, acceptance
      criteria, PR labels).
- [ ] `minion-proposal` issue filed on this repo, body carrying: the
      defect description with the manual-commit misattribution
      scenario; the session-recorded-PID approach; a conservative
      fallback when no PID was recorded (fall back to current global
      detection rather than regressing capture); the dead
      condensed-session block's removal or revival; and the program
      file marker.
- [ ] The issue's acceptance criteria explicitly require unit tests
      for the liveness decision (PID alive, PID dead, PID absent) and
      preservation of the PR #586 behavior: a live session's second
      commit still gets its trailer.
- [ ] Issue body is buildable standalone: a fresh minion session given
      only the issue can locate the hooks, session state, and detector
      seams from the descriptions in the body.

## Modules touched

- session-liveness proposal (this PRD); the eventual implementation
  touches the hooks and session-manager modules, but that lands via
  the pipeline, not this slice.

## Test prior art

- Proposal format: recent `minion-proposal` issues on this repo and
  the propose program's issue-creation steps — mirror their structure
  so the parser-facing parts (labels, marker comment) match exactly.

## Out of scope

- Implementing the liveness change (later minion work via approval).
- Any change to agent detection for attribution percentages —
  strictly the session-end/skip liveness decision.
