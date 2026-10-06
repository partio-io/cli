package review

import (
	"math"
	"slices"
	"strings"
	"unicode"
)

// stopWords carry no idea of their own in a proposal title.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true, "not": true, "no": true,
	"of": true, "to": true, "in": true, "on": true, "at": true, "by": true, "for": true, "from": true,
	"with": true, "without": true, "into": true, "onto": true, "as": true, "via": true, "per": true,
	"is": true, "are": true, "be": true, "was": true, "it": true, "its": true, "this": true, "that": true,
	"when": true, "if": true, "than": true, "then": true, "so": true, "all": true, "any": true, "each": true,
	"add": true, "support": true, "new": true, "partio": true,
}

// titleWords returns the set of words that carry the idea of a title:
// lower case, with apostrophes and hyphens joined, other punctuation
// as a break, stop words dropped, and a plural "s" cut.
func titleWords(title string) map[string]bool {
	title = strings.NewReplacer("'", "", "’", "", "-", "").Replace(strings.ToLower(title))
	words := map[string]bool{}
	for _, w := range strings.FieldsFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if stopWords[w] {
			continue
		}
		if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			w = strings.TrimSuffix(w, "s")
		}
		words[w] = true
	}
	return words
}

// sharedWords returns the words of a that b also has, in a stable order.
func sharedWords(a, b map[string]bool) []string {
	var out []string
	for w := range a {
		if b[w] {
			out = append(out, w)
		}
	}
	slices.Sort(out)
	return out
}

// titleMatch is the weighted overlap at or above which two titles are
// a strong match. On the backlog of 2026-10-06 (619 proposals) it keeps
// the confirmed pair #710/#736 (0.69) and yields about 0.4 candidates
// for each issue.
const titleMatch = 0.6

// wordWeights weighs each title word by its rarity in a backlog of
// titles, so that words most proposals carry ("checkpoint", "session",
// "agent") count for little.
type wordWeights struct {
	n  int
	df map[string]int
}

func newWordWeights(titles []map[string]bool) wordWeights {
	ww := wordWeights{n: len(titles), df: map[string]int{}}
	for _, words := range titles {
		for w := range words {
			ww.df[w]++
		}
	}
	return ww
}

func (ww wordWeights) weight(w string) float64 {
	return math.Log(float64(ww.n+1)/float64(ww.df[w]+1)) + 1
}

func (ww wordWeights) sum(words map[string]bool) float64 {
	var s float64
	for w := range words {
		s += ww.weight(w)
	}
	return s
}

// overlap returns the weight of the words a and b share over the weight
// of the smaller of the two, from 0 to 1, and the shared words.
func (ww wordWeights) overlap(a, b map[string]bool) (float64, []string) {
	shared := sharedWords(a, b)
	den := math.Min(ww.sum(a), ww.sum(b))
	if len(shared) == 0 || den == 0 {
		return 0, nil
	}
	var num float64
	for _, w := range shared {
		num += ww.weight(w)
	}
	return num / den, shared
}
