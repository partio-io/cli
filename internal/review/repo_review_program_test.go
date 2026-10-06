package review

import (
	"os"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/premise"
)

// reviewProgram is the review program, relative to the package directory
// that go test runs in.
const reviewProgram = "../../.minions/programs/review.md"

// TestReviewProgramTeachesAShapeTheGateAccepts holds the body skeleton of the
// review program to the gate's own check. The program once described a claim
// line with no backticks around its evidence, and a body with no heading after
// its claims. The gate refused both, and the first dry run on #31 ended with no
// verdict although the review chose the right outcome.
func TestReviewProgramTeachesAShapeTheGateAccepts(t *testing.T) {
	raw, err := os.ReadFile(reviewProgram)
	if err != nil {
		t.Fatalf("read the review program: %v", err)
	}
	skeleton, ok := fencedBlock(string(raw), premise.Marker)
	if !ok {
		t.Fatalf("%s shows no fenced body skeleton with %q", reviewProgram, premise.Marker)
	}
	if err := checkShape("", Rewrite{Title: "A title", Body: skeleton}); err != nil {
		t.Errorf("the body skeleton in %s fails the gate: %v", reviewProgram, err)
	}
}

// fencedBlock returns the first fenced block of src that contains want,
// without its fence lines and without the indent of its opening fence, as the
// model writes it out.
func fencedBlock(src, want string) (string, bool) {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		fence := strings.TrimLeft(lines[i], " ")
		if !strings.HasPrefix(fence, "```") {
			continue
		}
		indent := lines[i][:len(lines[i])-len(fence)]
		for j := i + 1; j < len(lines); j++ {
			if !strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
				continue
			}
			body := make([]string, 0, j-i-1)
			for _, line := range lines[i+1 : j] {
				body = append(body, strings.TrimPrefix(line, indent))
			}
			if block := strings.Join(body, "\n"); strings.Contains(block, want) {
				return block, true
			}
			i = j
			break
		}
	}
	return "", false
}
