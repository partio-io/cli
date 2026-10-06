# 08 — Nightly sweep

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [06 — Sweep acts on keep and close](./06-sweep-acts-on-keep-and-close.md)

## What to build

The sweep runs by itself at night, picks its own issues, and stops
when the operator switches it off.

**Picking the batch** (`review` package, `next` subcommand of
`minion-review`):

- List the open `minion-proposal` issues and drop those that carry
  `minion-reviewed`. GitHub's list filter cannot exclude a label, so
  filter in Go.
- Order: issues with `minion-approved` first, then by filing date,
  oldest first.
- Skip an issue that has two "no verdict" rows in the tracking issue,
  and list it under a "needs you" heading in the tracking issue body.
- `--sample N`: N issues spread across the filing months (take one per
  month in turn, oldest first in each month). The same input gives the
  same sample. The operator uses it for the 20-issue dry run.
- Print one issue number per line, for the workflow's loop.

**Tracking issue totals.** The tracking issue body shows totals per
verdict and per close reason, and the "needs you" list. The gate
recomputes them from the night comments after each row, so the body
cannot drift.

**The sweep workflow** gets:

- a schedule at 23:00 UTC
- a scheduled run acts only when the repository variable
  `PROPOSAL_REVIEW_SWEEP` is `on`; otherwise it ends at once. The
  operator switches it on after reading the dry run, and off to stop.
- a scheduled run takes no new issue after 05:00 UTC. That keeps it
  inside 00:00–07:00 local time in summer (UTC+2) and in winter
  (UTC+1). A manual dispatch is the operator's own choice and has no
  window.
- a dispatch input for a sample size, next to the issue list and the
  dry-run flag
- a job timeout that covers the window plus the last issue
- a concurrency group, so that two sweep runs never overlap

The runner runs one job at a time, and the sweep loops over its issues
inside one job. So GitHub's 24-hour limit on a queued job never
applies.

## User stories covered

1, 22, 25, 26, 27, 28, 29 (totals), 30, 31, 55.

## Acceptance criteria

- [ ] `next` returns open proposals without `minion-reviewed`, approved ones first, then by filing date, oldest first.
- [ ] `next` skips an issue with two "no verdict" rows and lists it under "needs you" in the tracking issue body.
- [ ] `next --sample N` returns N issues spread across filing months, and the same input gives the same sample.
- [ ] The tracking issue body shows totals per verdict and per close reason, recomputed after each row.
- [ ] The sweep workflow runs on a schedule at 23:00 UTC.
- [ ] A scheduled run acts only when the repository variable `PROPOSAL_REVIEW_SWEEP` is `on`, and otherwise ends at once.
- [ ] A scheduled run takes no new issue after 05:00 UTC, and a manual dispatch has no window.
- [ ] A manual dispatch accepts a sample size as well as an issue list and the dry-run flag.
- [ ] Two sweep runs never run at the same time.
- [ ] Tests for `next` and the totals run against a fake GitHub server, and repo tests pin the schedule, the variable check and the window.
- [ ] `make test` and `make lint` pass.

## Modules touched

- `review` package
- `minion-review` command
- sweep workflow
- program-shape repo tests

## Test prior art

- `internal/auditgate/run_test.go` — fake GitHub over `httptest`
- `internal/premise/repo_*_test.go` and
  `internal/programshape/repo_workflows_test.go` — tests that read a
  workflow file and pin its triggers and steps
- The propose workflow — the current scheduled workflow on the same
  runner

## Out of scope

- The build path (slices 09 and 10).
- The proposer (slice 11).
