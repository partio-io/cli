# 09 — Review before each build

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [07 — Sweep acts on rewrites](./07-sweep-acts-on-rewrites.md)

## What to build

Every build reviews its issue first, with the same program and gate as
the sweep. The review replaces the premise gate.

**The build workflow** (the minion build workflow that `minion-approved`
and `/minion build` start). Today its first step after the install runs
the premise-gate program, then reads the `do-not-build` label back to
decide "blocked". That gate fails open: no label means the build runs.
Replace that step with two:

1. Run the review program for the issue, with `MINION_REVIEW_DIR` set.
   A crashed session must not end the job before the gate runs (use the
   dead-code audit workflow's pattern).
2. Run `go run ./cmd/minion-review gate --mode build --issue N
   --verdict …`. In build mode the gate writes two step outputs to
   `$GITHUB_OUTPUT`: `blocked` and `changed`. Keep the output name
   `blocked`, because the later steps already test
   `steps.gate.outputs.blocked`.

| Verdict | Gate action | blocked | changed |
|---|---|---|---|
| keep | keep label rules, evidence comment | false | false |
| rewrite | rewrite as in slice 07 | false | true |
| close | close as in slice 06 | true | — |
| no verdict | exit 1 | — | — |

- **No verdict** fails the job, and the current failure step marks the
  issue. The build fails closed.
- **A close** is a block, not a failure: neither the done step nor the
  failure step runs, and the gate's comment explains the close.
- **A rewrite** makes the old research plan void, because the plan
  describes the old text. When `changed` is true, the workflow runs
  research even if the issue carries a slice plan. The step that
  confirms a plan then accepts only a plan comment published after the
  rewrite. The runtime reads the newest PRD and plan comments, so the
  new plan wins.

A build-time verdict writes its row to the tracking issue, marked as a
build.

**The premise-gate program is removed.** Update the program-set test
and the workflow tests. Build and chain tests that name the
premise-gate program now name the review step.

Until slice 10 lands, the gate treats every issue as a proposal, so it
could close or rewrite an issue that the operator wrote. Slices 09 and
10 ship in the same pull request.

## User stories covered

32, 33, 34, 35, 51, 52, 54.

## Acceptance criteria

- [x] The build workflow runs the review program and then `minion-review gate --mode build` as its first steps after the install, and no step runs the premise-gate program.
- [x] Build mode writes the `blocked` and `changed` outputs per the table above.
- [x] No verdict fails the job, and the current failure step marks the issue.
- [x] A close stops the run as blocked: neither the done step nor the failure step runs.
- [x] When `changed` is true, the workflow runs research even if the issue carries a slice plan.
- [x] After a rewrite, the build accepts only a slice plan published after the rewrite; an older plan does not count.
- [x] A build-time verdict writes its row to the tracking issue, marked as a build.
- [x] The premise-gate program is removed, and the program-set and workflow tests agree.
- [x] Build and chain tests that named the premise-gate program now name the review step, and they still prove that a blocked build creates no branch and no pull request.
- [x] Build-mode tests run against a fake GitHub server and assert the outputs for each outcome and for no verdict.
- [x] `make test` and `make lint` pass.

## Modules touched

- build workflow
- `review` package
- `minion-review` command
- premise-gate program (removed)
- premise repo tests
- program-shape repo tests

## Test prior art

- `internal/premise/repo_build_test.go` and `repo_chain_test.go` —
  step-order tests on the build workflow, the blocked path, and the
  research chain
- `internal/programshape/repo_workflows_test.go` — every program has a
  workflow or a recorded reason
- The dead-code audit workflow — `continue-on-error` on the session
  step, so that the gate fails closed

## Out of scope

- Facts-only mode for the operator's issues and for reopened issues
  (slice 10).
- The research program's own stop behaviour, which stays
  prompt-driven.
