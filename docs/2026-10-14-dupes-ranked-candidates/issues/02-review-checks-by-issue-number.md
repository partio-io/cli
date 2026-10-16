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

- [x] An issue-number query takes the title and the source items of the issue from the listing, and leaves the issue out of the candidates. A test shows a source match that comes only from the body of the issue.
- [x] When the listing does not hold the issue, the search reads it once (one request to `/repos/<repo>/issues/<n>`), and then runs as for a listed issue.
- [x] `dupes --issue <n>` prints the candidates of that issue as JSON. `--issue` with `--title`, `--source` or `--exclude` exits 2, and the message names the bad mix. Without `--issue`, a call with no `--title` still exits 2.
- [x] The duplicate step of the review program runs `go run ./cmd/minion-review dupes --issue <issue number>`, with no `--source`, `--title` or `--exclude`.
- [x] The step names the three kinds of match (`source`, `title`, `ranked`) and the fields of a candidate, `score` and `excerpt` among them.
- [x] The step tells the agent to read the title and the excerpt of each candidate, and to open with `gh issue view` only a candidate that can be the same idea.
- [x] The step judges a ranked candidate that is the same idea by the current rules, and it keeps the rule for a candidate that closed as a duplicate.
- [x] The step says: when not sure, keep the issue, with `"same": false` and the reason in `why`. Every candidate stays in the verdict.
- [x] The repo tests of the review program (`internal/premise/repo_review_dupes_test.go`) assert the points above. The current assertions pass; they change only where the text changes on purpose (`--issue` replaces `--source`, `--title` and `--exclude`).
- [x] A real dry run of the review program prints the new duplicate step in the agent instructions, and the text of steps 6, 7 and 8 after it. See "Dry run" below.
- [x] Live check, read-only: `go run ./cmd/minion-review dupes --issue 771` prints #750 among the candidates, and `--issue 758` prints #710.
- [x] No file under `.github/` changes, and `make test` and `make lint` pass.

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

## Evidence

- 1 red: TestWithIssueKeepsTheBodyOnTheQuery — dupes_test.go:345: withIssue = {... Body: ...}, want the title, the body and the exclusion of #758
- 2 red: TestFindDupesByIssueReadsAnUnlistedIssueOnce — dupes_test.go:330: read issues [], want [800]
- 1 green: TestWithIssueKeepsTheBodyOnTheQuery, TestFindDupesByIssueTakesTheIssueFromTheListing — ok  github.com/partio-io/cli/internal/review
- 2 green: TestFindDupesByIssueReadsAnUnlistedIssueOnce — ok  github.com/partio-io/cli/internal/review
- 3 red: TestDupesByIssuePrintsTheCandidatesOfThatIssue — dupes exit 2: flag provided but not defined: -issue
- 3 green: TestDupesByIssuePrintsTheCandidatesOfThatIssue, TestDupesByIssueRefusesTheOtherFlags, TestDupesNeedsATitle — ok  github.com/partio-io/cli/cmd/minion-review
- 4 red: TestReviewSearchesForDuplicates — repo_review_dupes_test.go:26: ### reviewer runs "go run ./cmd/minion-review dupes --source ..." want "... dupes --issue <issue number>"
- 5 red: TestReviewReadsTheRankedCandidates — repo_review_dupes_test.go:108: ### reviewer does not carry "is one of the 10 other proposals with the best score"
- 6 red: TestReviewReadsTheRankedCandidates — repo_review_dupes_test.go:107: ### reviewer does not carry "read the title and the excerpt of each candidate"
- 7 red: TestReviewReadsTheRankedCandidates — repo_review_dupes_test.go:107: ### reviewer does not carry "judge a ranked candidate that is the same idea as a strong match"
- 8 red: TestReviewReadsTheRankedCandidates — repo_review_dupes_test.go:107: ### reviewer does not carry "when you are not sure that a candidate is the same idea, keep the issue"
- 4 green: TestReviewSearchesForDuplicates — ok  github.com/partio-io/cli/internal/premise
- 5 green: TestReviewReadsTheRankedCandidates — ok  github.com/partio-io/cli/internal/premise
- 6 green: TestReviewReadsTheRankedCandidates — ok  github.com/partio-io/cli/internal/premise
- 8 green: TestReviewReadsTheRankedCandidates — ok  github.com/partio-io/cli/internal/premise
- 7 green: TestReviewReadsTheRankedCandidates, TestReviewAppliesTheDuplicateRules — ok  github.com/partio-io/cli/internal/premise
- 9 red: TestReviewSearchesForDuplicates — repo_review_dupes_test.go:30: ### reviewer still names --source: --issue replaces it
- 9 green: TestReviewSearchesForDuplicates, TestReviewRecordsEveryCandidate, TestReviewAppliesTheDuplicateRules, TestReviewReadsTheRankedCandidates — ok  github.com/partio-io/cli/internal/premise
- 10 smoke: minions v0.0.14 run review.md --issue 772 --dry-run — exit 0; agent instructions hold step 5 with 'dupes --issue <issue number>' (line 156), then 6 Decide (193), 7 Compose (214), 8 Write the verdict (263)
- 11 smoke: live: minion-review dupes --issue 771 / --issue 758 — 771: #750 ranked second (score 0.599); 758: #710 ranked second (score 0.45)
- 12 smoke: git status .github; make test; make lint — .github unchanged; make test exit 0; golangci-lint 0 issues

## Report

**What changed.**

- `internal/review/dupes.go`: `DupesQuery` has `Issue` and `Body`. With `Issue`, `withIssue` takes the title and the body of the issue from the listing, or from one `GetIssue` when the listing does not hold it. It adds the body to the sources and sets `Exclude`. It clones `Sources`, so the slice of the caller does not change.
- `cmd/minion-review/main.go`: `dupes --issue <n>`. A mix with `--title`, `--source` or `--exclude` exits 2, and the message names the bad flags. A negative `--issue` exits 2. `dupesUsage` and the package comment are updated.
- `.minions/programs/review.md` step 5: runs `dupes --issue <issue number>`, names the three kinds of match and the fields, reads the title and the excerpt first, judges a ranked candidate as a strong match, and keeps the issue with `"same": false` when not sure.
- Tests: 3 in `internal/review/dupes_test.go` (the fake backlog now serves `/issues/{n}` and counts the reads), 2 in `cmd/minion-review/main_test.go` (with a `captureStderr` helper), 1 new and 1 changed in `internal/premise/repo_review_dupes_test.go`.

**Suite.** `make test`, `make lint` and `go vet ./...`: green at step 1 and at step 5 (344 tests pass, 0 fail, lint 0 issues). No repair outside "Modules touched". No file under `.github/` changed.

**Evidence.** 9 `red`, 9 `green`, 3 `smoke` (a rerun of a criterion replaces its record). Criteria 10 to 12 are smoke checks: the dry run (minions v0.0.14, `--issue 772`) printed the new step 5, then steps 6, 7 and 8; the live run put #750 second for #771 and #710 second for #758.

**Review findings** (`code-review medium --fix`):

1. Whole body gives source items, so a "Related:" reference becomes a source match. Skipped: the criteria ask for body source items. Ask item.
2. `GetIssue` accepted a pull request number. Fix removed: no criterion asks for it. Ask item.
3. Negative `--issue` was not a usage error. Fixed, kept: a bug in the flag rules of criterion 3.
4. `DupesQuery.Body` was not read. The review removed the field; argos put it back, because the issue says "keep the body on the query" for slice 03.
5. `withIssue` changed the `Sources` array of the caller. Fixed, kept (clone). The comment says that `Issue` replaces `Title`, `Body` and `Exclude`.
6. `captureStderr` did not restore `os.Stderr` on a panic. Fixed, kept (`defer`).
7. The test refuses `--source`, `--title`, `--exclude` anywhere in the reviewer text. Skipped: intended. Ask item.

**Spec findings:**

- (a) The body was not on the query. Criterion 1 went through red → green again (`TestWithIssueKeepsTheBodyOnTheQuery`).
- (b) The pull-request guard. Removed (see review finding 2).
- (c) Step 5 said "best title score"; the issue says "best score", which stays true after slice 03. Criterion 5 went through red → green again.

**Marked files.** None: the slice created no new file.

**Check time.** 08:04:30 to 08:05:53, about 1.5 minutes.

**Surprise.** Live, #750 and #708 score 0.5992 for #771, just below the title cut of 0.6. They come back as `ranked`, not `title`. Ask item.
