package review

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

// The recall floors: the counts of the first run over the pairs of the
// analysis of 2026-10-13. A model of the search found 30 of 42
// hand-found pairs and 77 of 78 tool-found pairs.
const (
	recallFloorHand = 30
	recallFloorTool = 77
)

// recallIssue is one proposal of the corpus, in the form of
// `gh issue list --json`.
type recallIssue struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`
	StateReason string `json:"stateReason"`
	Body        string `json:"body"`
}

func readRecallCorpus(t *testing.T) []recallIssue {
	t.Helper()
	f, err := os.Open("testdata/recall/corpus.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var corpus []recallIssue
	if err := json.NewDecoder(zr).Decode(&corpus); err != nil {
		t.Fatal(err)
	}
	return corpus
}

// readRecallPairs reads one pair per line: the closed issue, then the
// kept issue.
func readRecallPairs(t *testing.T, file string) [][2]int {
	t.Helper()
	f, err := os.Open("testdata/recall/" + file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var out [][2]int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var closed, kept int
		if _, err := fmt.Sscan(sc.Text(), &closed, &kept); err == nil {
			out = append(out, [2]int{closed, kept})
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// recallBase is the corpus in the REST form.
func recallBase(corpus []recallIssue) []fakeIssue {
	out := make([]fakeIssue, len(corpus))
	for i, is := range corpus {
		out[i] = fakeProposal(is.Number, is.Title, is.Body, strings.ToLower(is.State), strings.ToLower(is.StateReason))
	}
	return out
}

// recallBacklog is the base backlog, with the pre-merge body of the kept
// issue when the analysis used one.
func recallBacklog(t *testing.T, base []fakeIssue, closed, kept int) []fakeIssue {
	t.Helper()
	before, err := os.ReadFile(fmt.Sprintf("testdata/recall/premerge/%d.body.before-%d.md", kept, closed))
	if errors.Is(err, fs.ErrNotExist) {
		return base
	}
	if err != nil {
		t.Fatal(err)
	}
	out := slices.Clone(base)
	for i := range out {
		if out[i].Number == kept {
			out[i].Body = string(before)
		}
	}
	return out
}

func TestFindDupesRecallOfTheKnownCopies(t *testing.T) {
	base := recallBase(readRecallCorpus(t))
	lists := []struct {
		name, file string
		floor      int
		found, ran atomic.Int32
		total      int
	}{
		{name: "hand", file: "pairs-github-dup.tsv", floor: recallFloorHand},
		{name: "tool", file: "pairs-tool.txt", floor: recallFloorTool},
	}
	t.Run("pairs", func(t *testing.T) {
		for i := range lists {
			l := &lists[i]
			pairs := readRecallPairs(t, l.file)
			l.total = len(pairs)
			for _, p := range pairs {
				closed, kept := p[0], p[1]
				t.Run(fmt.Sprintf("%s/%d-%d", l.name, closed, kept), func(t *testing.T) {
					t.Parallel()
					l.ran.Add(1)
					f := newFakeBacklog(t, recallBacklog(t, base, closed, kept))
					got, err := FindDupes(f.client(), DupesQuery{Repo: "partio-io/cli", Issue: closed})
					if err != nil {
						t.Fatal(err)
					}
					for _, d := range got {
						if d.Number == kept {
							l.found.Add(1)
							return
						}
					}
				})
			}
		}
	})
	if lists[0].total != 42 || lists[1].total != 78 {
		t.Errorf("read %d hand-found and %d tool-found pairs, want 42 and 78", lists[0].total, lists[1].total)
	}
	// A -run filter that selects some pairs only skips the floors.
	var low, partial bool
	for i := range lists {
		low = low || int(lists[i].found.Load()) < lists[i].floor
		partial = partial || int(lists[i].ran.Load()) < lists[i].total
	}
	if partial {
		t.Logf("recall: hand-found %d of %d run, tool-found %d of %d run; floors not checked",
			lists[0].found.Load(), lists[0].ran.Load(), lists[1].found.Load(), lists[1].ran.Load())
		return
	}
	if low {
		t.Errorf("recall below the floor: hand-found %d of %d (floor %d), tool-found %d of %d (floor %d)",
			lists[0].found.Load(), lists[0].total, lists[0].floor,
			lists[1].found.Load(), lists[1].total, lists[1].floor)
	}
	t.Logf("recall: hand-found %d of %d, tool-found %d of %d",
		lists[0].found.Load(), lists[0].total, lists[1].found.Load(), lists[1].total)
}
