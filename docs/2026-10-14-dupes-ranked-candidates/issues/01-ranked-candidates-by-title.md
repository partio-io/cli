# 01 — Ranked candidates by title

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

`minion-review dupes` returns the best candidates, not only the
candidates over a cutoff.

Today `FindDupes` (`internal/review/dupes.go`) returns a candidate only
for a strong match: a proposal whose body cites a source item of the
query (`"match": "source"`), or a proposal with a title overlap of 0.6
or more (`"match": "title"`, `titleMatch` in `title_words.go`). #771 had
a title overlap of 0.599 with #750, so the search returned nothing, and
the proposer filed a third copy of one idea.

After this slice:

- The strong matches stay as they are, in the same order: source
  matches first, then title matches.
- After them, the search returns up to 10 ranked candidates
  (`"match": "ranked"`, a new `Match*` constant): the proposals with
  the best title score above zero that are not strong matches. The
  title score is the weighted overlap that `wordWeights.overlap`
  computes today. For a ranked candidate, `shared` holds the shared
  title words.
- Equal scores sort by issue number, older first: the older issue is
  the one that most likely stays.
- The weight sums run in sorted order, so that the same input gives
  the same score on every run. Today `wordWeights.sum` follows the map
  order.
- Each candidate, strong or ranked, gets two new fields: `score` (a
  number, its title score) and `excerpt` (a string). Slice 03 adds the
  code names to the score.
- The excerpt is the start of the description of the candidate as
  plain text: about 300 characters, cut at the end of a word, with no
  heading marks and no backslash escapes. On 2026-10-13, 375 of the 640
  proposals started with `## Description`, 183 with `## Summary`, and
  32 held escaped backticks (`` \` ``). So the excerpt of a body that
  starts with `## Description` starts with the first words under that
  heading.
- The JSON output of the command keeps its current fields and adds
  `score` and `excerpt`. The flags do not change in this slice.

The agent reads the title and the excerpt of each candidate, and it
opens an issue only when the candidate can be the same idea. So the
excerpt saves one tool call for each candidate that is clearly another
idea.

A model of this search (title score only) on the live backlog of
2026-10-14 put #708 and #750 at the top of the ranked part for the
title of #771. Both have a title overlap of 0.5999, just under the
cutoff.

## User stories covered

1, 2, 3, 4, 6 (for title words), 8 (by title), 9, 10, 11, 26.

## Acceptance criteria

- [ ] The strong matches come first, as today and in the same order. A strong match does not come again in the ranked part.
- [ ] After the strong matches, the search returns at most 10 ranked candidates, with `"match": "ranked"`, highest score first. For each one, `shared` holds the title words that it shares with the query.
- [ ] A proposal with a score of zero is not returned as a ranked candidate.
- [ ] Equal scores put the older issue (the lower number) first.
- [ ] Two runs over the same issues give the same candidates, with the same scores, in the same order. The weight sums run in sorted order.
- [ ] A query with no source item (a changelog entry) still gets ranked candidates by its title.
- [ ] Each candidate, strong or ranked, has a `score` and an `excerpt`.
- [ ] The excerpt is the start of the description as plain text: no heading marks, no backslash escapes, about 300 characters, and it stops at the end of a word.
- [ ] The JSON of `minion-review dupes` keeps every current field and adds `score` (a number) and `excerpt` (a string). The command test checks both new fields.
- [ ] The current tests in `internal/review/dupes_test.go` and `cmd/minion-review/main_test.go` pass. They change only where a ranked candidate or a new field now shows, for example the "Compact checkpoint session" case, which now gets ranked candidates.
- [ ] Live check, read-only: `GH_TOKEN=$(gh auth token) GITHUB_REPOSITORY=partio-io/cli go run ./cmd/minion-review dupes --title "Respect CLAUDE_CONFIG_DIR in Claude session discovery" --source entireio/cli#2627 --exclude 771` prints #750 and #708 among the candidates. On 2026-10-14, neither one was a strong match for this query.
- [ ] `make test` and `make lint` pass.

## Modules touched

- Weights (`internal/review/title_words.go`: the sums in sorted order)
- Duplicate search (`internal/review/dupes.go`: the ranked part and the new fields; a new file for the excerpt, one concern per file)
- The `dupes` command (`cmd/minion-review/main.go`: the JSON output and the package comment)

## Test prior art

- `internal/review/dupes_test.go`: `newFakeBacklog`, `fakeProposal` and `fillers` serve a backlog through the fake list endpoint. Test through `FindDupes`, and assert the candidates and their fields, not the score math.
- `internal/review/testdata/issue-*-body.md`: real bodies for the excerpt cases (a body with `## Description`, a body with escapes).
- `cmd/minion-review/main_test.go`, `TestDupesPrintsTheCandidatesAsJSON`: the command test, with an `httptest` server and `t.Setenv`.

## Out of scope

- The code names and their weight table: [03 — Code names in the score](./03-code-names-in-the-score.md).
- `--issue` and the review program: [02 — The review checks by issue number](./02-review-checks-by-issue-number.md).
- `--body-file` (slice 03) and the proposer program (slices 04 and 05).
- The recall test on the 120 labeled pairs: slice 06.
- A match by equal proposal id (PRD, Out of Scope).
