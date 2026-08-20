package premise

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFindLocatesTheSectionWithoutGuessing(t *testing.T) {
	const claim = "- A claim. [evidence: `a.go`]"

	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "the section ends the body",
			body: "Description.\n\n" + Heading + "\n\n" + Marker + "\n\n" + claim + "\n",
			want: true,
		},
		{
			name: "the section sits between two other sections",
			body: "## Why\n\nreasons\n\n" + Heading + "\n\n" + Marker + "\n\n" + claim +
				"\n\n## Acceptance criteria\n\n- something else\n",
			want: true,
		},
		{
			name: "the body carries no premise section",
			body: "## Why\n\nreasons\n",
			want: false,
		},
		{
			name: "a hand-written heading with no marker is not the block",
			body: "## Premise\n\nWe assume the hook walks the whole tree.\n",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Find(tt.body)
			if ok != tt.want {
				t.Fatalf("Find found = %v, want %v; section %q", ok, tt.want, got)
			}
			if !ok {
				return
			}
			if !strings.Contains(got, claim) {
				t.Errorf("section is missing its claim: %q", got)
			}
			if strings.Contains(got, "Acceptance criteria") {
				t.Errorf("section runs past the heading that ends it: %q", got)
			}
		})
	}
}

// issueBody wraps a premise section in the prose an issue carries around it, so
// the tests exercise the section boundaries and not just the claim lines.
func issueBody(claims string) string {
	return "Some description of the proposal.\n\n" +
		Heading + "\n\n" + Marker + "\n\n" + claims + "\n\n" +
		"## Next steps\n\nprose that is not a claim\n"
}

func TestParseReadsOneClaimPerLine(t *testing.T) {
	tests := []struct {
		name    string
		claims  string
		want    []Claim
		wantErr bool
	}{
		{
			name:   "a claim names the path that proves it",
			claims: "- The post-commit hook reads the commit diff. [evidence: `internal/hooks/post_commit.go`]",
			want: []Claim{{
				Statement: "The post-commit hook reads the commit diff.",
				Evidence:  "internal/hooks/post_commit.go",
			}},
		},
		{
			name:   "a claim may name a command instead of a path",
			claims: "- Attribution is binary. [evidence: `go doc ./internal/attribution`]",
			want: []Claim{{
				Statement: "Attribution is binary.",
				Evidence:  "go doc ./internal/attribution",
			}},
		},
		{
			name: "claims keep the order they were written in",
			claims: "- First claim. [evidence: `a.go`]\n" +
				"- Second claim. [evidence: `b.go`]",
			want: []Claim{
				{Statement: "First claim.", Evidence: "a.go"},
				{Statement: "Second claim.", Evidence: "b.go"},
			},
		},
		{
			name:    "a claim continued on a second line is not one claim",
			claims:  "- The hook walks the whole tree\n  on every commit. [evidence: `internal/hooks/post_commit.go`]",
			wantErr: true,
		},
		{
			name:    "prose in the section is not a claim",
			claims:  "This proposal assumes the hook is slow. [evidence: `internal/hooks/post_commit.go`]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(issueBody(tt.claims))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse succeeded, want an error; got %+v", got.Claims)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(got.Claims, tt.want) {
				t.Errorf("Parse claims =\n%+v\nwant\n%+v", got.Claims, tt.want)
			}
		})
	}
}

func TestParseRejectsAClaimWithNoEvidence(t *testing.T) {
	tests := []struct {
		name   string
		claims string
	}{
		{
			name:   "no evidence field at all",
			claims: "- The post-commit hook walks the whole tree.",
		},
		{
			name:   "an empty evidence field",
			claims: "- The post-commit hook walks the whole tree. [evidence: ``]",
		},
		{
			name:   "an evidence field of only spaces",
			claims: "- The post-commit hook walks the whole tree. [evidence: `   `]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(issueBody(tt.claims))
			if !errors.Is(err, ErrNoEvidence) {
				t.Fatalf("Parse error = %v, want ErrNoEvidence; got claims %+v", err, got.Claims)
			}
		})
	}
}

func TestRenderRejectsAClaimWithNoEvidence(t *testing.T) {
	tests := []struct {
		name  string
		block Block
		want  error
	}{
		{
			name:  "no claims",
			block: Block{},
			want:  ErrNoClaims,
		},
		{
			name:  "a claim with no evidence",
			block: Block{Claims: []Claim{{Statement: "The hook walks the whole tree."}}},
			want:  ErrNoEvidence,
		},
		{
			name:  "a claim with blank evidence",
			block: Block{Claims: []Claim{{Statement: "The hook walks the whole tree.", Evidence: "  "}}},
			want:  ErrNoEvidence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.block)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Render error = %v, want %v; rendered %q", err, tt.want, got)
			}
			if got != "" {
				t.Errorf("Render returned %q with an error, want the empty string", got)
			}
		})
	}
}

func TestRenderRoundTripsThroughParse(t *testing.T) {
	want := Block{Claims: []Claim{
		{Statement: "The post-commit hook reads the commit diff.", Evidence: "internal/hooks/post_commit.go"},
		{Statement: "Attribution is binary.", Evidence: "go doc ./internal/attribution"},
	}}

	section, err := Render(want)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	got, err := Parse("Some description.\n\n" + section)
	if err != nil {
		t.Fatalf("Parse of a rendered section: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip =\n%+v\nwant\n%+v", got, want)
	}
}
