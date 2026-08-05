# Minion PR Audits

## Problem Statement

Minion-built PRs ship with quality defects that today only a careful
human review catches. The review of PR #586 surfaced three recurring
classes:

- **Semantically dead code** — blocks that are still called but can no
  longer do anything (a hardcoded argument makes a branch unreachable).
  No deterministic tool flags these: the code is "used", so `unused`
  and `deadcode` see it as live. Only reasoning catches it.
- **Silently swallowed errors** — error returns discarded with a blank
  identifier, so a failure path degrades behavior with no trace. The
  repo's lint config currently permits this.
- **Missing end-to-end coverage** — a change whose headline behavior is
  only verified at the unit level, with nothing driving the full flow.

Each defect class currently depends on the operator reading every diff
line. That does not scale with the minion pipeline's throughput, and a
defect that slips through merges silently.

Separately, the same review found that agent-liveness detection is
machine-global: the hooks ask "is any process with 'claude' in its
command line running anywhere on this machine?" rather than "is the
agent that owns this repo's session alive?". On a workstation where
some agent is nearly always running, a manual commit can be rewritten
with agent-attribution trailers and a redundant checkpoint — wrong data
in the product whose core promise is accurate AI attribution.

## Solution

Give the pipeline its own reviewers, matched to the nature of each
defect class:

- **Deterministic class → lint gate.** Enable errcheck's blank-discard
  option and fix the existing violations. From then on, every minion
  build and every local run fails on a swallowed error — no new
  workflow, no judgment, no noise.
- **Judgment classes → two PR-time audit actions.** Each minion-labeled
  PR gets two automated reviewers, each an independent GitHub Action
  running a minions prompt program on the self-hosted minion runner:
  - The **dead-code audit** reviews the PR diff for semantically dead
    code. Findings trigger exactly one fix round on the PR branch
    itself, followed by one re-audit. If the branch comes back clean,
    the check is green and the PR the operator reviews is already
    fixed. If findings survive, the check goes red with a comment
    listing them — and it never loops.
  - The **e2e-need audit** judges whether the change warrants an
    end-to-end test. When it does, it files a `minion-proposal` issue
    with acceptance criteria a fresh minion session can build from —
    feeding the work back into the queue the operator already triages.
    It never turns the PR red on findings.
- **Attribution correctness → one proposal now.** File a
  `minion-proposal` issue to replace machine-global process-name
  liveness with a check of the session's own recorded agent process
  ID. The implementation then flows through the normal
  approve → build → PR pipeline.

## User Stories

1. As the pipeline operator, I want minion PRs automatically audited for dead code before I review them, so that called-but-inert blocks never reach main.
2. As the pipeline operator, I want the dead-code audit to fix its findings on the PR branch itself, so that the PR I review is already clean instead of me round-tripping feedback.
3. As the pipeline operator, I want the fix round strictly bounded to one attempt plus one re-audit, so that automated sessions can never ping-pong on a PR.
4. As the pipeline operator, I want a red check plus one comment listing surviving findings when the fix round does not fully clean up, so that I know exactly what still needs a human call.
5. As the pipeline operator, I want a quiet green check when an audit finds nothing, so that clean PRs gather no comment noise.
6. As the pipeline operator, I want both audits to run only on minion-labeled PRs, so that my hand-written PRs are not taxed with agent sessions.
7. As the pipeline operator, I want audits scoped to the PR's diff plus whatever surrounding code the session chooses to read, so that findings are attributable to the change under review rather than repo-wide archaeology.
8. As the pipeline operator, I want an e2e-need audit that files a proposal issue when a change deserves end-to-end coverage, so that test debt enters the same queue I already triage.
9. As the pipeline operator, I want the e2e-need audit to never turn a PR red on findings, so that missing future coverage does not block shipping the fix in front of me.
10. As the pipeline operator, I want e2e proposals deduplicated against open proposals before filing, so that re-runs do not flood the queue with copies.
11. As the pipeline operator, I want e2e proposals to carry acceptance criteria concrete enough for a fresh minion session, so that approving one is all I have to do to get the test built.
12. As the pipeline operator, I want blank-discarded errors caught by lint on every build, so that a swallowed failure like the liveness check's can never merge silently again.
13. As the pipeline operator, I want the existing blank-discard violations fixed in the same change that enables the rule, so that the gate starts green instead of permanently red.
14. As the pipeline operator, I want test files exempt from the blank-discard rule, so that deliberate discards in tests stay idiomatic.
15. As the pipeline operator, I want every audit comment to start with a fixed marker prefix, so that I can filter machine comments by body the way the rest of the pipeline already does.
16. As the pipeline operator, I want the audit's own fix commits recognizable and ignored by audit triggers, so that an audit push can never re-trigger itself into a loop.
17. As the pipeline operator, I want a dry-run mode on both audit workflows, so that I can see exactly what prompt would run without a PR being touched.
18. As the pipeline operator, I want each audit proven on a staged scratch PR before it watches real PRs, so that the loop's behavior is verified the same way the slice pipeline was.
19. As the pipeline operator, I want an escape-hatch label that makes both audits skip a PR, so that I can move urgently without editing workflows.
20. As the pipeline operator, I want a crashed audit (one that produced no verdict) to fail visibly even for the never-blocking audit, so that an infra failure cannot masquerade as a pass.
21. As the pipeline operator, I want both audit workflows pinned to the same minions version scheme as the existing workflows, so that upgrades stay one pin-bump PR across the board.
22. As a reviewer of minion PRs, I want surviving findings phrased with location and reasoning, so that I can judge each one without re-deriving the analysis.
23. As the pipeline operator, I want a session-scoped liveness proposal filed with clear acceptance criteria, so that the machine-global process grep gets replaced through the normal minion flow rather than a hand-patched hotfix.
24. As a partio user, I want my manual commits never stamped with agent attribution just because some agent is running elsewhere on my machine, so that attribution data stays truthful.
25. As the pipeline operator, I want re-running an audit on an unchanged PR to converge to the same verdict without duplicate comments or duplicate proposals, so that retries are safe.

## Implementation Decisions

- **Everything ships in the cli repo.** No changes to the minions
  runtime or binary — this avoids the release-tag-plus-pin-bump deploy
  chain entirely. The audit programs are prompt programs executed by
  the already-pinned minions version.
- **Two separate actions, not one combined audit** (operator's call
  during research): independent cadence, independent evolution, at the
  cost of two sessions per PR.
- **Trigger:** pull-request events (opened, labeled, synchronize) gated
  on the PR carrying the minion label. A skip label short-circuits both
  audits. Both run on the existing self-hosted minion runner.
- **Loop guard:** the dead-code audit's fix commits carry a fixed
  commit-subject prefix; a guard step exits early when the PR head
  commit carries it. Concurrency groups cancel superseded audit runs on
  the same PR.
- **Verdict contract (shared by both audits):** the audit session must
  write a small verdict file — status plus a findings list — as its
  final act. A deterministic workflow step turns the verdict into the
  check outcome. A missing or malformed verdict fails the workflow for
  both audits (fail-closed on infra); a well-formed verdict with
  findings is red only for the dead-code audit.
- **Dead-code fix round:** on findings, the same session (or an
  immediately spawned fresh one) applies fixes, commits with the marker
  prefix, pushes to the PR branch, and re-audits once. Findings that
  survive go into the verdict and one PR comment. Hard bound: one
  round, no exceptions.
- **E2e-need flow:** findings become one `minion-proposal` issue per
  distinct gap, deduplicated by searching open proposals first, each
  carrying a program-file reference and acceptance criteria, matching
  the propose program's issue format. The audit itself always reports a
  passing verdict when it ran to completion.
- **Comment authorship:** comments and issues are authored via the
  pipeline's existing PAT and therefore appear as the operator;
  machine-filtering relies on the fixed body prefix, never on author —
  consistent with the pipeline's existing convention.
- **Lint tightening:** enable errcheck's blank-discard option in the
  repo lint config with test files excluded; fix all current
  violations (36 at time of writing) in the same change. The minion's
  per-build check gate and local runs pick the rule up with no further
  wiring.
- **Session-scoped liveness proposal (F1):** one `minion-proposal`
  issue proposing that session-end and skip decisions consult the
  liveness of the session's recorded agent process ID instead of a
  machine-wide process-name grep, with a conservative fallback when no
  PID was recorded. Implementation is deliberately left to the minion
  pipeline; this PRD only files the proposal.
- **Language:** the cli repo is Go and stays Go for any code touched;
  the new surface area is workflow YAML and markdown prompt programs,
  so no new-language decision arises.

## Testing Decisions

- Good tests assert external behavior — what a module produces given
  inputs — never internal wiring. For this feature the external
  behaviors are: the verdict file's effect on check status, the
  presence/absence and content shape of comments and proposals, the
  loop guard's short-circuit, and lint failing on a blank-discarded
  error.
- **Every module is tested.** Concretely:
  - The lint change is tested by the lint itself plus the existing
    suite staying green after the violation fixes.
  - The verdict gate and loop guard are exercised with fixture verdicts
    (pass, fail, missing, malformed) and a marker-prefixed head commit
    in a staged run before the actions watch real PRs.
  - Each audit program is validated first via the workflows' dry-run
    path (prompt rendered, nothing touched), then via a staged scratch
    PR — the same staged-run pattern the slice pipeline was verified
    with (prior art: the per-slice staged smoke test).
  - The F1 proposal's acceptance criteria require unit tests for the
    liveness decision when it is implemented; that testing obligation
    ships inside the proposal text.
- **Deliberate gap, named:** the quality of the LLM's judgment inside
  an audit session (does it spot real dead code, does it over-flag) is
  not unit-testable and is not tested in CI. Mitigations: the strict
  verdict schema, the one-round bound, staged runs before enablement,
  and the operator reviewing every comment and proposal it produces.

## Out of Scope

- Whole-repo or scheduled audit sweeps — audits look at PR diffs only
  (the scheduled-sweep option was explicitly declined).
- Auditing human-authored PRs.
- Wiring the auto-approve program to a schedule; proposals filed by the
  e2e audit wait for the operator's label like every other proposal.
- Any change to the minions runtime, its release, or its version pins
  beyond adding the two workflows at the current pin.
- Implementing the session-scoped liveness fix (only the proposal is in
  scope here).
- Extending any of this to repos other than cli.
- Branch protection or required-check enforcement; check status stays
  advisory signal for the operator.

## Further Notes

- Origin: findings F1–F4 from the operator review of PR #586
  (2026-07-28). F2 (dead code), F3 (swallowed error), and F4 (missing
  e2e) become the three loops above; F1 becomes the filed proposal.
- Expected cost envelope: two audit sessions per minion PR, comparable
  to the implement session itself (order of a dollar per PR at current
  usage); dry-run and the skip label are the relief valves if this
  proves noisy.
- If PR-time auditing proves high-signal, a later iteration may add a
  scheduled sweep for drift in code that predates the audits — kept out
  of scope now by explicit decision.
- The known EnsureRepos stale-checkout backlog item is unrelated and
  untouched by this work.
