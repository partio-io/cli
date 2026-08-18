package premise

// A proposal filed before the premise block existed states its facts in prose.
// There are hundreds of them, and 56 already carry the approved label, so a
// gate that only reads blocks leaves every one of them exempt. Where the claims
// come from is decided here, once, so no stage invents its own answer.

// Source names where a stage takes the claims it verifies.
type Source string

const (
	// SourceBlock reports a proposal that carries a premise block. The block
	// is the claims, and the prose around it is not read.
	SourceBlock Source = "block"

	// SourceProse reports a proposal with no premise block. The claims are
	// extracted from the issue prose at check time. Nothing is written back:
	// the body keeps its own words and gains no block.
	SourceProse Source = "prose"
)

// ClaimSource reports where the claims for an issue body come from. A body that
// carries a block uses the block, so a proposal that states its premise is
// never re-read from its prose.
func ClaimSource(body string) Source {
	if _, ok := Find(body); ok {
		return SourceBlock
	}
	return SourceProse
}
