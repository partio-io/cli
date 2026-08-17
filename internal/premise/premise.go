// Package premise defines the premise block that a filed proposal carries: the
// factual claims the proposal depends on, each paired with the evidence that
// settles it. The block is machine-checkable, so a later stage rechecks a
// proposal from this block alone and never has to parse the surrounding prose.
package premise

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// Heading starts the premise section of an issue body.
	Heading = "## Premise"

	// Marker names the block and its format version, so a later stage finds
	// the section without guessing at the prose around it.
	Marker = "<!-- partio:premise:v1 -->"
)

var (
	// ErrNotFound reports a body that carries no premise section.
	ErrNotFound = errors.New("premise: no premise section")

	// ErrNoClaims reports a premise section that states nothing.
	ErrNoClaims = errors.New("premise: section carries no claims")

	// ErrNoEvidence reports a claim with no evidence to settle it.
	ErrNoEvidence = errors.New("premise: claim has no evidence")
)

// Claim is one factual statement about the repository, with the evidence that
// settles it: a path, a symbol, or a command whose output decides the matter.
type Claim struct {
	Statement string
	Evidence  string
}

// Block is the set of claims a proposal depends on.
type Block struct {
	Claims []Claim
}

// claimLine matches one claim. The whole claim is on one line, and the evidence
// closes it, so a claim that names nothing cannot pass as prose.
var claimLine = regexp.MustCompile("^-\\s+(.*\\S)\\s+\\[evidence:\\s*`([^`]*)`\\]$")

// Render writes a block as the premise section a proposal carries. It refuses a
// claim with no evidence rather than emit an empty field for a later stage to
// puzzle over.
func Render(b Block) (string, error) {
	if len(b.Claims) == 0 {
		return "", ErrNoClaims
	}

	var out strings.Builder
	out.WriteString(Heading + "\n\n" + Marker + "\n\n")

	for _, c := range b.Claims {
		statement := strings.TrimSpace(c.Statement)
		evidence := strings.TrimSpace(c.Evidence)

		if statement == "" {
			return "", fmt.Errorf("premise: claim states nothing (evidence %q)", evidence)
		}
		if evidence == "" {
			return "", fmt.Errorf("%w: %s", ErrNoEvidence, statement)
		}
		if strings.ContainsAny(statement, "\n\r") || strings.ContainsAny(evidence, "\n\r`") {
			return "", fmt.Errorf("premise: claim does not fit one line: %s", statement)
		}

		fmt.Fprintf(&out, "- %s [evidence: `%s`]\n", statement, evidence)
	}

	return out.String(), nil
}

// Find returns the premise section of an issue body, from the heading to the
// next heading of the same level or to the end of the body. It reports false
// unless the section also carries Marker.
func Find(body string) (string, bool) {
	lines := strings.Split(body, "\n")

	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == Heading {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", false
	}

	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			end = i
			break
		}
	}

	section := strings.Join(lines[start:end], "\n")
	if !strings.Contains(section, Marker) {
		return "", false
	}
	return section, true
}

// Parse reads the premise block out of an issue body. It rejects a claim that
// attaches no evidence rather than return it with an empty field.
func Parse(body string) (Block, error) {
	section, ok := Find(body)
	if !ok {
		return Block{}, ErrNotFound
	}

	var b Block
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == Marker {
			continue
		}

		m := claimLine.FindStringSubmatch(line)
		if m == nil {
			if !strings.Contains(line, "[evidence:") {
				return Block{}, fmt.Errorf("%w: %s", ErrNoEvidence, line)
			}
			return Block{}, fmt.Errorf("premise: line is not one claim: %s", line)
		}
		if strings.TrimSpace(m[2]) == "" {
			return Block{}, fmt.Errorf("%w: %s", ErrNoEvidence, m[1])
		}

		b.Claims = append(b.Claims, Claim{
			Statement: m[1],
			Evidence:  strings.TrimSpace(m[2]),
		})
	}

	if len(b.Claims) == 0 {
		return Block{}, ErrNoClaims
	}
	return b, nil
}
