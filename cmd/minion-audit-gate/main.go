// Command minion-audit-gate maps a minion audit verdict file to a CI
// outcome. It exits 0 only for a valid pass verdict; anything else —
// fail verdict, missing or malformed file — exits 1 (fail-closed)
// after upserting the single audit comment on the PR.
//
// Given the repair round it belongs to, the gate also reports a spent
// repair budget: on the last round, surviving findings produce a
// comment that says the pipeline gave up rather than one that reads
// like a first-pass failure.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/partio-io/cli/internal/auditgate"
	"github.com/partio-io/cli/internal/repairround"
)

func main() {
	var (
		pr      = flag.Int("pr", 0, "pull request number (required)")
		audit   = flag.String("audit", "", "audit name for the comment, e.g. dead-code (required)")
		verdict = flag.String("verdict", ".minion-audit/verdict.json", "path to the verdict file")
		round   = flag.Int("round", 0, "repair round this run belongs to; 0 means no round accounting")
		// The default is the same constant minion-repair-round counts
		// against, so the two sides of one round cannot disagree about
		// the cap unless a caller overrides both.
		maxRounds = flag.Int("max", repairround.DefaultMaxRounds, "repair rounds allowed for this check")
	)
	flag.Parse()
	if *pr == 0 || *audit == "" {
		fmt.Fprintln(os.Stderr,
			"usage: minion-audit-gate --pr <number> --audit <name> [--verdict <path>] [--round <n>] [--max <n>]")
		os.Exit(2)
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	token := os.Getenv("GH_TOKEN")
	if repo == "" || token == "" {
		fmt.Fprintln(os.Stderr, "minion-audit-gate: GITHUB_REPOSITORY and GH_TOKEN must be set")
		os.Exit(2)
	}
	api := os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}

	res, err := auditgate.Run(auditgate.Config{
		VerdictPath: *verdict,
		Repo:        repo,
		PRNumber:    *pr,
		AuditName:   *audit,
		APIBaseURL:  api,
		Token:       token,
		Round:       *round,
		MaxRounds:   *maxRounds,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-audit-gate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("minion-audit-gate:", res.Reason)
	if !res.Green {
		os.Exit(1)
	}
}
