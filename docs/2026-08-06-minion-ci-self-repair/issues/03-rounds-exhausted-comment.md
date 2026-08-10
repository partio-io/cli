# 03 — Say when it gave up

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Count repair rounds instead of forbidding them](./02-count-repair-rounds.md)

## What to build

With round counting in place, a pull request can exhaust its repair
budget and stop. From the outside that is indistinguishable from a
check that never attempted a repair at all — both look like a red
check with a findings comment. Make the difference visible.

When the budget is spent and findings still survive, the check's
comment says so: it names the findings that remain and states how many
repair rounds ran. The operator should be able to act on it without
opening a workflow log.

The comment surface already exists and already upserts — a pull request
accumulates one comment per check rather than one per run. That
behavior must hold: three rounds must not leave three comments, and the
final comment replaces whatever the earlier rounds left.

## User stories covered

PRD user stories 15, 16, 17, 18.

## Acceptance criteria

- [x] When repair rounds are exhausted and findings survive, the
      check's pull request comment names the surviving findings.
- [x] That comment states how many repair rounds ran.
- [x] The comment is distinguishable from a first-round failure comment,
      so "gave up after spending the budget" reads differently from
      "failed on the first pass".
- [x] Repeated rounds upsert a single comment per check rather than
      appending a new one each round.
- [x] The pull request is left red and open. Nothing closes it, rebuilds
      it, or re-runs the build.
- [x] Tests cover the rounds-exhausted comment body and the upsert path
      that replaces a prior comment rather than appending.

## Modules touched

`internal/auditgate` (modified), per the PRD's Implementation Decisions.
Consumes the round number produced by `internal/repairround` from slice
[02](./02-count-repair-rounds.md).

## Test prior art

`internal/auditgate/run_test.go` is the direct precedent and the file
this slice extends: table-driven, standard library `testing` only, a
local `httptest` server standing in for the GitHub API so no test
touches the network, and `t.TempDir()` for verdict-file fixtures.

Assert on the comment body's observable content and on which HTTP verb
the upsert chose — not on unexported helper names or call ordering.

## Out of scope

- Rebuilding or closing an exhausted pull request. The PRD rules that
  out explicitly; the pipeline stops and reports.
- Changing what the audits themselves find or how findings are worded.
- Notifying anywhere other than the pull request comment — no email, no
  Slack, no issue.
- The lint/test gate's comments, which arrive with slice
  [05](./05-lint-test-pr-check.md) and reuse this same surface.
