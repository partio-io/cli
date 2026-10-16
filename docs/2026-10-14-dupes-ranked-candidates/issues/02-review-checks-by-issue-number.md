# 02 — The review checks by issue number

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — Ranked candidates by title](./01-ranked-candidates-by-title.md)

## What to build

The review gives the duplicate check only the number of the issue
under review. The tool reads the rest itself, and the review program
judges the ranked candidates of slice 01.

**The search.** `DupesQuery` gets an issue number. With it, `FindDupes`
takes the title, the body and the source items (`SourceRefs` of the
body) of that issue from the proposal listing that it reads anyway, and
it leaves that issue out of the candidates. When the listing does not
hold the issue (a build of an issue without the `minion-proposal`
label), the search reads it once with `github.Client.GetIssue`. Keep
the body on the query: slice 03 takes its code names.

**The command.** `minion-review dupes --issue <n>`. `--issue` takes no
`--title`, `--source` or `--exclude`: a mix exits 2, with a message
that names the flags. (`--body-file` comes in slice 03, which adds it
to this rule.) Without `--issue`, `--title` stays required. Update
`dupesUsage` and the package comment.

**The review program** (`.minions/programs/review.md`, step 5, "Look
for duplicates"):

- Run `go run ./cmd/minion-review dupes --issue <issue number>`, with
  no other flag.
- Name the three kinds of match: `"match": "source"` (cites the same
  source item), `"match": "title"` (a title overlap of 0.6 or more) and
  `"match": "ranked"` (one of the 10 other proposals with the best
  score). Say that each candidate carries its `title`, `state`,
  `state_reason`, `labels`, `score`, `shared` and `excerpt`.
- Read the title and the excerpt of each candidate. Open a candidate
  with `gh issue view <n>` only when it can be the same idea.
- Judge a ranked candidate as a strong match: the current rules apply
  ("The first issue reviewed stays"). Keep the current rule for a
  candidate that closed as a duplicate: it stands for the issue that it
  names.
- When the agent is not sure that a candidate is the same idea, it
  keeps the issue: `"same": false`, with the reason in `why`. No issue
  closes on a weak guess.
- Every candidate stays in the verdict, with `same` and `why`, as
  today. Keep "The list is the backlog you judge against".

The text stays in the `## Agents` prose of the `### reviewer` agent,
with bold labels and no `####` heading: the minions parser drops every
other H2 section and every `####` heading. A static test cannot prove
that the text reaches the agent; only a real dry run can.

The gate does not change. The verdict keeps its form, and the gate does
not compare the verdict with the candidate list. The workflows do not
change either: they read the program from `main` and run the command
with `go run`, so the merge is live at the next run.

A model of the search on the live backlog of 2026-10-14, with the
title score only, put #750 at the top for #771, and #710 second for
#758.

## User stories covered

2, 20, 21, 22, 23, 24, 25 (for `--issue`), 30 (for the review program), 31 (for the review).

## Acceptance criteria

- [ ] An issue-number query takes the title and the source items of the issue from the listing, and leaves the issue out of the candidates. A test shows a source match that comes only from the body of the issue.
- [ ] When the listing does not hold the issue, the search reads it once (one request to `/repos/<repo>/issues/<n>`), and then runs as for a listed issue.
- [ ] `dupes --issue <n>` prints the candidates of that issue as JSON. `--issue` with `--title`, `--source` or `--exclude` exits 2, and the message names the bad mix. Without `--issue`, a call with no `--title` still exits 2.
- [ ] The duplicate step of the review program runs `go run ./cmd/minion-review dupes --issue <issue number>`, with no `--source`, `--title` or `--exclude`.
- [ ] The step names the three kinds of match (`source`, `title`, `ranked`) and the fields of a candidate, `score` and `excerpt` among them.
- [ ] The step tells the agent to read the title and the excerpt of each candidate, and to open with `gh issue view` only a candidate that can be the same idea.
- [ ] The step judges a ranked candidate that is the same idea by the current rules, and it keeps the rule for a candidate that closed as a duplicate.
- [ ] The step says: when not sure, keep the issue, with `"same": false` and the reason in `why`. Every candidate stays in the verdict.
- [ ] The repo tests of the review program (`internal/premise/repo_review_dupes_test.go`) assert the points above. The current assertions pass; they change only where the text changes on purpose (`--issue` replaces `--source`, `--title` and `--exclude`).
- [ ] A real dry run of the review program prints the new duplicate step in the agent instructions, and the text of steps 6, 7 and 8 after it. See "Dry run" below.
- [ ] Live check, read-only: `go run ./cmd/minion-review dupes --issue 771` prints #750 among the candidates, and `--issue 758` prints #710.
- [ ] No file under `.github/` changes, and `make test` and `make lint` pass.

## Dry run

Install minions v0.0.14 into the scratchpad of the session, then run
the program from the repository, for an open proposal (for example
#772):

```bash
GOBIN=<scratchpad>/bin go install github.com/partio-io/minions/cmd/minions@v0.0.14
cd /home/arvos/Code/workspaces/workspace-minion-reboot/cli
GH_TOKEN=$(gh auth token) WORKSPACE_ROOT=/home/arvos/Code/workspaces/workspace-minion-reboot \
  <scratchpad>/bin/minions run .minions/programs/review.md --issue 772 --dry-run --project .minions/project.yaml
```

The workflow runs the program with `--issue "$ISSUE"`
(`.github/workflows/proposal-review.yml`). In the output, check the
agent instructions: the new step 5 text, then the text of steps 6, 7
and 8.

## Modules touched

- Duplicate search (`internal/review/dupes.go`: the issue-number query)
- The `dupes` command (`cmd/minion-review/main.go`: `--issue`, the flag rules, the usage)
- The review program (`.minions/programs/review.md`)

## Test prior art

- `internal/review/dupes_test.go`: `newFakeBacklog` serves only the list endpoint today. Add the issue endpoint (`GET /repos/partio-io/cli/issues/{n}`) and count its requests.
- `cmd/minion-review/main_test.go`: `TestDupesNeedsATitle`, and the usage-error tests of `gate` and `check`.
- `internal/premise/repo_review_dupes_test.go`: `TestReviewSearchesForDuplicates`, `TestReviewRecordsEveryCandidate` and `TestReviewAppliesTheDuplicateRules`, with the helpers `reviewer` and `flat`.

## Out of scope

- The code names and `--body-file`: [03 — Code names in the score](./03-code-names-in-the-score.md).
- The proposer program: slices 04 and 05.
- The gate and the form of the verdict (PRD: no gate change).
- The e2e-audit program: it does not use this check.
