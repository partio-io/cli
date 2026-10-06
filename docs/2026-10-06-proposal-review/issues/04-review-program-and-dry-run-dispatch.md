# 04 — Review program and dry-run dispatch

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Catch assumed facts in premise checks](./02-catch-assumed-facts.md), [03 — Review verdict and dry-run gate](./03-verdict-and-dry-run-gate.md)

## What to build

The judgment half of the review, and the first end-to-end path: the
operator dispatches a workflow with a list of issues, a fresh session
judges each one, and the gate from slice 03 records each verdict in the
tracking issue. Nothing changes on the reviewed issues.

**The review program** (new, next to the other programs that the
workflows run). One agent, one fresh session per issue, run with
`minions run <program> --issue N`. The runtime gives the agent the
issue body under an "Issue" section. The agent:

1. Reads the source item that the issue cites (an entireio/cli pull
   request, issue or changelog entry) with read-only `gh` calls.
2. Applies the verifier to the premise against the checked-out tree:
   the `## Premise` block when the issue has one, else the prose, with
   the assumed-fact rule from slice 02.
3. Judges fit with the ingest prompt's relevance rule (Git workflows,
   AI agent sessions, code attribution, checkpoints).
4. Judges whether the idea is already built, from the tree.
5. Decides keep, rewrite or close, with one close reason: built, false
   premise, does not apply, or could not verify. Duplicates arrive in
   slice 05.
6. For a keep or a rewrite, composes the body in the issue shape that
   today's proposer files: what to build in full, `## Acceptance
   Criteria` as a `- [ ]` checklist, a `## Premise` section with the
   `<!-- partio:premise:v1 -->` marker and one claim per line with
   `[evidence: …]`, the gathered evidence, and a final `Proposal id:
   <id>` line. The rewrite keeps the original idea and its source
   link, and it drops the `<!-- program: … -->` pointer to a proposal
   file that no longer exists.
7. Writes the verdict JSON (contract in slice 03) to
   `$MINION_REVIEW_DIR/verdict.json`.

The agent makes no change on GitHub. It writes nothing in its working
directory: a file there is read as build output and turns the run into
a pull request. Its instructions sit in its `## Agents` section as
prose with bold labels, because the runtime drops `## Steps` sections
and `####` headings.

**The sweep workflow** (new). In this slice it has manual dispatch only,
with two inputs: `issues` (a list of issue numbers) and `dry_run`
(default true; only true works until slice 06). One job on the
self-hosted runner `github-runner-partio-minion-ai-01`, with
`GH_TOKEN: ${{ secrets.GH_PAT }}`. It installs minions at the same
pinned version as the other workflows. For each issue, in order, it
clears the review directory, runs the review program with
`MINION_REVIEW_DIR` set to an absolute path in the workspace, then runs
`go run ./cmd/minion-review gate --issue N --verdict … --dry-run
--night <UTC date set once per run>`. A crashed session or a "no
verdict" exit does not stop the loop.

**The verifier** loses two passages: "The backlog is not swept…" and
"Do not rewrite the issue body. Do not backfill a premise block…". In
their place: the verifier writes nothing itself, and what a caller
writes with the result belongs to the caller. The premise-gate program
and the stage gate keep their own no-backfill text, so the gates do
not change behavior.

Note: GitHub dispatches only a workflow that exists on the default
branch, so the first real dry run happens after merge. Use #30 and #31
as known-answer cases then. Both are closed, and the dispatch accepts
closed issues, because a dry run writes nothing to them.

## User stories covered

2, 3, 13 (program side), 14, 17, 24, 49.

## Acceptance criteria

- [ ] A review program exists with one agent whose instructions sit in its agents section, and the program-shape test passes.
- [ ] The program applies the verifier and the ingest prompt's relevance rule, and it describes the proposer's issue shape for a keep or a rewrite.
- [ ] The program writes only the verdict file at `$MINION_REVIEW_DIR/verdict.json`, and it says that it makes no GitHub change and writes nothing in its working directory.
- [ ] The program keeps the original idea and source link in a rewrite, and drops the pointer to a proposal file.
- [ ] The sweep workflow has a manual dispatch with an issue list and a dry-run flag, and it runs the program and then the gate for each issue, in sequence, in one job.
- [ ] A crashed session or a "no verdict" exit does not stop the loop.
- [ ] The workflow installs minions at the same pinned version as the other workflows, and it runs the gate with `go run`.
- [ ] The verifier no longer says that the backlog is not swept or forbids a rewrite, while the premise-gate program and the stage gate still forbid a backfill.
- [ ] The August verifier tests that pinned "not swept" and "not rewritten" now pin the split: the verifier checks, and the caller writes.
- [ ] A repo test proves that the workflow runs the program before the gate for each issue, and the program and workflow tests know the new program.
- [ ] `make test` and `make lint` pass.

## Modules touched

- review program (new)
- sweep workflow (new)
- verifier
- premise repo tests
- program-shape repo tests

## Test prior art

- `internal/premise/repo_*_test.go` — step-finding helpers that read a
  workflow and check step order; the verifier tests that this slice
  rewrites
- `internal/programshape/repo_*_test.go` — the list of programs and
  the workflows that run them
- The dead-code audit workflow — an absolute verdict directory in the
  workspace, and a session step that must not end the job before a
  fail-closed gate runs

## Out of scope

- Duplicate judgment (slice 05).
- Any change to a reviewed issue (slices 06 and 07).
- The schedule, `next` and the sample input (slice 08).
- The build path (slice 09).
