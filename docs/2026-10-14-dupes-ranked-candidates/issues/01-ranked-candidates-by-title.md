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

- [x] The strong matches come first, as today and in the same order. A strong match does not come again in the ranked part.
- [x] After the strong matches, the search returns at most 10 ranked candidates, with `"match": "ranked"`, highest score first. For each one, `shared` holds the title words that it shares with the query.
- [x] A proposal with a score of zero is not returned as a ranked candidate.
- [x] Equal scores put the older issue (the lower number) first.
- [x] Two runs over the same issues give the same candidates, with the same scores, in the same order. The weight sums run in sorted order.
- [x] A query with no source item (a changelog entry) still gets ranked candidates by its title.
- [x] Each candidate, strong or ranked, has a `score` and an `excerpt`.
- [x] The excerpt is the start of the description as plain text: no heading marks, no backslash escapes, about 300 characters, and it stops at the end of a word.
- [x] The JSON of `minion-review dupes` keeps every current field and adds `score` (a number) and `excerpt` (a string). The command test checks both new fields.
- [x] The current tests in `internal/review/dupes_test.go` and `cmd/minion-review/main_test.go` pass. They change only where a ranked candidate or a new field now shows, for example the "Compact checkpoint session" case, which now gets ranked candidates.
- [x] Live check, read-only: `GH_TOKEN=$(gh auth token) GITHUB_REPOSITORY=partio-io/cli go run ./cmd/minion-review dupes --title "Respect CLAUDE_CONFIG_DIR in Claude session discovery" --source entireio/cli#2627 --exclude 771` prints #750 and #708 among the candidates. On 2026-10-14, neither one was a strong match for this query.
- [x] `make test` and `make lint` pass.

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

## Evidence

- 1 red: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — dupes_ranked_test.go:39:17: undefined: MatchRanked
- 1 green: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — ok  	github.com/partio-io/cli/internal/review
- 2 red: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — dupes_ranked_test.go:39:17: undefined: MatchRanked
- 2 green: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — ok  	github.com/partio-io/cli/internal/review
- 3 red: TestFindDupesRanksNoCandidateWithAZeroScore — dupes_ranked_test.go:90: FindDupes = [{Number:700 ... {Number:1 Title:Filler idea 1 ... Score:0 ...], want only #700 with a score above zero (guard set to score >= 0)
- 3 green: TestFindDupesRanksNoCandidateWithAZeroScore — ok  	github.com/partio-io/cli/internal/review
- 4 red: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — dupes_ranked_test.go:51: ranked [701 703 705 707 709 711 704 706 708 710], want the coding titles first, then the older plugin titles
- 4 green: TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches — ok  	github.com/partio-io/cli/internal/review
- 5 red: TestFindDupesGivesTheSameCandidatesOnEveryRun — dupes_ranked_test.go:77: a second run gave [... Score:0.6073330149619556 ...] want [... Score:0.6073330149619557 ...]
- 5 green: TestFindDupesGivesTheSameCandidatesOnEveryRun — ok  	github.com/partio-io/cli/internal/review (-count=5)
- 6 red: TestFindDupesRanksByTitleForAQueryWithNoSource — dupes_ranked_test.go:105: FindDupes = [], want #750 ranked (ranked part removed)
- 6 green: TestFindDupesRanksByTitleForAQueryWithNoSource — ok  	github.com/partio-io/cli/internal/review
- 7 red: TestFindDupesGivesEachCandidateAScoreAndAnExcerpt — dupes_ranked_test.go:135: #690 = "source", score 0.000, excerpt ""; want "source", scored false, excerpt "Origin: entireio/cli#2075. Rename the flag."
- 7 green: TestFindDupesGivesEachCandidateAScoreAndAnExcerpt — ok  	github.com/partio-io/cli/internal/review
- 8 red: TestFindDupesExcerptIsTheStartOfTheDescriptionAsPlainText — dupes_ranked_test.go:163: #710 excerpt = "## Description `installHooks` ...", want the first words under ## Description (heading drop and unescape removed)
- 8 green: TestFindDupesExcerptIsTheStartOfTheDescriptionAsPlainText — ok  	github.com/partio-io/cli/internal/review
- 9 red: TestDupesPrintsTheCandidatesAsJSON — main_test.go:51: dupes printed no "score" field (json tags set to "-")
- 9 green: TestDupesPrintsTheCandidatesAsJSON — ok  	github.com/partio-io/cli/cmd/minion-review
- 10 red: TestFindDupesMatchesAStrongTitleOverlap/Compact_checkpoint_session — dupes_test.go:166: #701 matched by "ranked", want "title"
- 10 green: go test ./internal/review/ ./cmd/minion-review/ — ok  	github.com/partio-io/cli/internal/review; ok  	github.com/partio-io/cli/cmd/minion-review
- 11 smoke: live minion-review dupes, partio-io/cli, --exclude 771 — exit 0; #708 ranked 0.5992, #750 ranked 0.5992 first and second of 10 ranked; no strong match
- 12 green: make test; make lint — make test exit 0; golangci-lint run: 0 issues.

## Report

**What changed**

- `internal/review/dupes.go`: a new `MatchRanked` constant and the `maxRanked = 10` cap. After the strong matches, `FindDupes` gives the best non-strong proposals with a title score above zero, highest score first, and a lower issue number first when scores are equal. Each `Dupe` has the new fields `Score` (`score`) and `Excerpt` (`excerpt`). After the review fix, the excerpt is made only for the 10 ranked candidates that the search keeps.
- `internal/review/title_words.go`: `wordWeights.sum` adds the weights in sorted word order. Before this change, two runs gave scores that were different in the last bit (0.6073330149619556 and 0.6073330149619557). So the tie order changed from run to run.
- `internal/review/excerpt.go` (new): `excerpt(body)` removes the ATX heading lines and the HTML comments, undoes the backslash escapes, joins the white space, and cuts at the end of a word at 300 runes or less.
- `cmd/minion-review/main.go`: only the package comment changes. The JSON gets the two new fields from the struct tags.
- Tests: `internal/review/dupes_ranked_test.go` (new, 6 tests), `internal/review/excerpt_test.go` (new, written by the review), and `TestDupesPrintsTheCandidatesAsJSON`, which now checks every field and the type of `score` and `excerpt`. The changes in `dupes_test.go`: the strong-match tests skip `ranked` entries, the "Compact checkpoint session" case now expects ranked candidates, and two `want` values get `Score` and `Excerpt`.

**Suite**: `make test` was green and `make lint` gave `0 issues` in step 1. They were green again in step 5, after the review fixes. I repaired no module outside "Modules touched".

**Evidence**: 10 `red`, 11 `green`, 1 `smoke`. The tracer test failed at compile time. Criteria 3, 6, 8 and 9 were already green from earlier code, so their red runs come from a temporary mutation of the code: the `>= 0` guard, the ranked append removed, the heading drop and the unescape removed, and the JSON tags set to `"-"`. Each record names its mutation.

**Live check (criterion 11)**: `go run ./cmd/minion-review dupes … --exclude 771` gave exit 0 and no strong match. #708 and #750 are ranked first and second, both at 0.5992. Then come #715 (0.51) and #144 (0.48).

**Review findings** (`code-review medium --fix`, 7 findings):

1. Every line that starts with `#` was dropped, also a text line such as "#771 is …". Fixed: only an ATX heading line is dropped. This is part of criterion 8, so the fix stays.
2. HTML comments such as `<!-- partio:premise:v1 -->` stayed in the excerpt. Fixed. The fix stays: criterion 8 asks for "plain text", and a marker is not text for a reader.
3. An excerpt was made for each ranked hit before the cut to 10. Fixed: the output is the same and less work is done. The fix stays.
4. `.minions/programs/review.md` and the propose program still say that the list holds only source and title matches, and tell the agent to open each candidate. Skipped: slices 02 and 04 change these programs. This is on the end list as an Ask.
5. The backslash escapes are also undone inside code spans. Skipped, on the end list as an Ask.
6. If the first word is longer than 300 runes, the excerpt cuts in the middle of it. Skipped, on the end list as an Ask.
7. The end-of-word check in the test reads the raw body, which has newlines. Skipped: the check is weak but correct for the current testdata. On the end list as an Ask.

**Spec findings**: (a) none. (c) none. (b) one item: the removal of the HTML comments. I kept it, for the reason in review finding 2. `excerpt_test.go` covers it, but the spec check did not see that file, because I marked it only after the check.

**Marked with `git add -N`**: `internal/review/excerpt.go`, `internal/review/dupes_ranked_test.go`, `internal/review/excerpt_test.go`.

**Check time**: the review and the spec check ran in parallel for 2 min 9 s (07:59:34 to 08:01:43).

**Note against the PRD**: until slices 02 and 04 merge, each review and each proposer run gets up to 10 more candidates. Their prompts do not explain the `ranked` match yet (review finding 4). So do not ship this slice alone without those slices.
