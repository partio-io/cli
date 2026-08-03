// Command minion-audit-gate maps a minion audit verdict file to a CI
// outcome. It exits 0 only for a valid pass verdict; anything else —
// fail verdict, missing or malformed file — exits 1 (fail-closed)
// after upserting the single audit comment on the PR.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/partio-io/cli/internal/auditgate"
)

func main() {
	var (
		pr      = flag.Int("pr", 0, "pull request number (required)")
		audit   = flag.String("audit", "", "audit name for the comment, e.g. dead-code (required)")
		verdict = flag.String("verdict", ".minion-audit/verdict.json", "path to the verdict file")
	)
	flag.Parse()
	if *pr == 0 || *audit == "" {
		fmt.Fprintln(os.Stderr, "usage: minion-audit-gate --pr <number> --audit <name> [--verdict <path>]")
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
