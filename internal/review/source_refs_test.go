package review

import (
	"slices"
	"testing"
)

func TestSourceRefsNormalizesEachForm(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"shorthand", "Inspired by entireio/cli#2237 (hook backup).", []string{"entireio/cli#2237"}},
		{"shorthand upper case", "see EntireIO/CLI#7", []string{"entireio/cli#7"}},
		{"issue URL", "Inspired by [x](https://github.com/entireio/cli/issues/1960).", []string{"entireio/cli#1960"}},
		{"pull request URL", "See https://github.com/entireio/cli/pull/2641/files", []string{"entireio/cli#2641"}},
		{"redirect URL", "**Origin:** [entireio/cli#11](https://redirect.github.com/entireio/cli/issues/11)", []string{"entireio/cli#11"}},
		{"Origin line", "## Source\n\n- **Origin:** entireio/cli#594 (changelog 0.5.0)\n", []string{"entireio/cli#594"}},
		{"Origin line with two refs", "**Origin:** entireio/cli#1, entireio/cli#2", []string{"entireio/cli#1", "entireio/cli#2"}},
		{"Origin line with a bare second number", "**Origin:** [entireio/cli#5](https://redirect.github.com/entireio/cli/issues/5) #6 (changelog 0.6)", []string{"entireio/cli#5", "entireio/cli#6"}},
		{"legacy source line", "<!-- minion-task\nid: x\nsource: 'entireio/cli#594 #601 (changelog 0.5.0)'\n-->", []string{"entireio/cli#594", "entireio/cli#601"}},
		{"repo then PR", "This mirrors the bug in entireio/cli PR #2641, adapted.", []string{"entireio/cli#2641"}},
		{"backticked repo then PRs", "Inspired by `entireio/cli` PRs #2075 and #2077 (Goose).", []string{"entireio/cli#2075", "entireio/cli#2077"}},
		{"repo then PR list", "Inspired by: entireio/cli PR #2075, #2076, #2077 (Goose).", []string{"entireio/cli#2075", "entireio/cli#2076", "entireio/cli#2077"}},
		{"repo then changelog then number", "**Source:** entireio/cli changelog 0.4.2 (#812)", []string{"entireio/cli#812"}},
		{"repo then parenthesis", "the approach in entireio/cli (changelog 0.5.0, PRs #1, #2).", []string{"entireio/cli#1", "entireio/cli#2"}},
		{"repo then numbers with a slash", "and entireio/cli #1520/#1521 where git hung", []string{"entireio/cli#1520", "entireio/cli#1521"}},
		{"repo then version", "Inspired by entireio/cli v0.5.1 `entire sessions` subcommands (#9, #10)", []string{"entireio/cli#9", "entireio/cli#10"}},
		{"repo then capitalized changelog", "**Source:** entireio/cli Changelog 0.4.2 (#812)", []string{"entireio/cli#812"}},
		{"repo then lower-case pr", "see entireio/cli pr #2641", []string{"entireio/cli#2641"}},
		{"changelog only", "Inspired by entireio/cli v0.5.1 changelog: `entire activity` command", nil},
		{"italic shorthand", "_Inspired by entireio/cli#1912_", []string{"entireio/cli#1912"}},
		{"bare number without a repo", "Fixes #31 in this repo.", nil},
		{"bare number on a later line", "Origin: entireio/cli#1\nSee #31 here.", []string{"entireio/cli#1"}},
		{"path is not a repo", "Look in `internal/git` for #12 and cli/internal/hooks/", nil},
		{"changelog URL is not an item", "https://github.com/entireio/cli/blob/main/CHANGELOG.md", nil},
		{"repeat counts once", "entireio/cli#3 and https://github.com/entireio/cli/issues/3", []string{"entireio/cli#3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SourceRefs(tt.text); !slices.Equal(got, tt.want) {
				t.Errorf("SourceRefs(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
