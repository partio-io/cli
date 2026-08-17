# 07 — The build refuses a stale premise

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — The proposer checks before it files](./04-proposer-checks-before-filing.md)

## What to build

Make the build stage verify the premise before it writes any code, and
refuse to open a pull request for work already known to be void.

This is the last gate and the most expensive one to skip. Issue #30
reached three builds and produced a pull request that improves nothing:
measured directly, the change it delivers is level with the original on
a thousand-file repository and roughly thirty to forty percent slower
on a ten-thousand-file one. Every minute of that was spent after the
premise had already been false for months.

The build stage reuses the verification behaviour from slice 4 and the
stop behaviour from slice 6 — the same label, the same comment shape,
the same refusal to close the issue. It checks against the tree it is
about to modify, which is the whole reason this repeats rather than
trusting the research stage's earlier verdict.

A holding premise records what was verified before the build proceeds,
so a pull request carries the evidence its work rests on.

## User stories covered

PRD stories 6, 7, 8, 9, 10, 11, 28, 29.

## Acceptance criteria

- [x] The build stage verifies the premise before writing any code
      — the premise is checked by its own program, `.minions/programs/premise-gate.md`,
      which `.github/workflows/minion.yml` runs in a step ahead of the one that
      runs `implement.md`. The tracer-bullet test compares step positions rather
      than positions in the file, because the build program's path is named in an
      earlier `Determine program` step than the step that runs it. Confirmed red
      first: the gate program did not exist.
- [x] Verification reuses the behaviour from slice 4
      — the gate program's `### premise-checker` names `.minions/premise-verifier.md`
      and applies it "as written", and the program carries neither `VerifierMarker`
      nor `GateMarker`, which is what a pasted copy looks like. Proved by mutation:
      dropping "as written" fails the test. A dry run shows both shared documents
      arriving in `## Pre-Read Context` in full, 7848 characters.
- [x] The stop behaviour matches slice 6 — same label, same comment shape
      — one test covers both programs at once: each must reference `GatePath`, and
      neither may carry `GateMarker`, `BlockingLabel` or `GateCommentMarker`. The
      label, the comment marker and the override all reach the agent from
      `stage-gate.md`, confirmed in the rendered dry run. `stage-gate.md` needed no
      edit: it already reads "The research and implement programs apply this
      description". Proved by mutation: naming the label in the program fails this
      test and the override test.
- [x] A failed premise opens no pull request and creates no branch
      — this is why the gate is its own program run by the workflow rather than an
      agent inside `implement.md`. The runtime creates the branch before the agent
      session starts (`slice.go:139`), commits `--allow-empty` for the slice marker
      (`:214`), then pushes (`:217`) and opens the PR (`:225`) with no guard, so a
      build stopped from inside still leaves a branch and a marker-only PR. The
      workflow skips the build step entirely instead. The gate's own run takes the
      non-slice path, where an unmodified worktree returns `Skipped` before any
      push (`executor.go:320-325`) — which is why the program forbids writing into
      the working directory, since `git status --porcelain` counts untracked files.
      Proved by mutation on all three: the guard, `slices: true`, and the rule.
- [x] A failed premise stops before any slice of the plan is executed, not partway
      through — the same guard, one layer up. `minions run implement.md` never
      executes, so slice 1 never starts and no worktree is made. Re-verification
      between slices stays out, as the issue asks: the check runs once, in its own
      workflow step.
- [x] The stage never closes the issue
      — the gate program carries neither `gh issue close` nor `minion-done`. The
      workflow does close the issue on a normal run, so the test also requires
      every step carrying `gh issue close`, `minion-done` or `minion-failed` to be
      guarded by the gate's verdict. Without that, a blocked run would have failed
      red on the `Mark done` step's "no PR exists" check and labelled the issue
      `minion-failed` — a blocked premise is neither done nor failed. Proved by
      mutation: unguarding `Mark done` fails the test.
- [x] Removing the label lets a subsequent run proceed
      — the gate program never names `do-not-build`, so it cannot skip on it and
      re-verifies against the tree on every run. The workflow reads the label only
      after the gate has run, which the test pins by comparing positions inside
      that step, so the verdict acted on is this run's rather than one left over.
      Proved by mutation: naming the label in the program fails the test.
- [x] A holding premise records what was verified, and the build proceeds unchanged
      — the checker records every claim, its evidence, its verdict and the excerpt
      behind it for a block that holds as well as one that does not, and refreshes
      the issue's premise section with what this run read. The build step carries
      exactly one condition, the blocked verdict and nothing else, which the test
      checks line by line so a second condition cannot be added quietly.
- [x] The implement program's instructions live where the parser carries them, and
      the shape test stays green — this was genuinely red, and it was a live defect
      rather than a formality: `## Slice builds` and `## PR and commit quality` sat
      under second-level headings the parser drops, so none of that prose ever
      reached the model. `programshape.Check` could not catch it, because it
      returns nil for any program that defines an `## Agents` section. Both sections
      now sit inside the `### implement` agent. The first attempt used `####`
      sub-headings and the dry run showed them dropped too — `splitSections` splits
      on every heading level and a level-4 section matches no case, so its body is
      discarded silently. Bold labels replaced them. The test now pins both rules,
      and `go test ./internal/programshape/` passes.
- [x] `make test` and `make lint` pass
      — both exit 0; `golangci-lint` reports 0 issues, and no package fails.
      `minions run cli/.minions/programs/implement.md --dry-run` renders `AGENTS: 1`
      and carries the whole instruction set into `## Agent Instructions`;
      `premise-gate.md` renders `AGENTS: 1` with both shared documents pre-read.

## Modules touched

`stage-gate`, `premise-verifier`, `implement-program`, `build-workflow`,
`premise-gate-program`.

The last two were added during execution, with the operator's decision,
because the first three cannot satisfy "opens no pull request and creates
no branch" on their own. The pinned runtime offers a program no way to
stop a slice build before it pushes: `skip_marker` is read only on the
non-slice path, `stage_files` and `depends_on` are parsed and never read,
unknown capability keys are ignored in silence, and the agent loop
continues past a failed agent by design. The check therefore runs as its
own program, `.minions/programs/premise-gate.md`, in a
`.github/workflows/minion.yml` step ahead of the build.

`stage-gate` and `premise-verifier` needed no edit. Both already name the
implement program as a caller, which is what "described once" was for.

## Test prior art

- Slice 4's verifier fixtures and slice 6's gate behaviour. This slice
  should not invent a second way to express either.
- The implement program already runs the repository's own checks
  between plan slices; the premise check belongs before the first
  slice, alongside the existing setup rather than inside the loop.

## Out of scope

- The research stage's check. That is slice 6. Neither stage depends on
  the other having run.
- Proposals with no premise block. That is slice 8.
- Any change to how the build slices a plan, resumes, or counts slice
  markers.
- Re-verifying between individual plan slices. One check before the
  build starts is what this slice delivers.

## Verification note

Verify the failure path without spending a real build: confirm that no
branch and no pull request appear, not merely that the run reports a
stop. A run reporting success while having done almost nothing is the
exact failure this whole PRD exists to prevent, so check the artefacts
rather than the run banner.
