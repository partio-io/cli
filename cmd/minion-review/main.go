// Command minion-review is the deterministic half of the proposal
// review. Its gate subcommand reads the verdict file the review program
// wrote for one issue, checks it, and records it as one row in the
// tracking issue. It fails closed: a missing or invalid verdict is "no
// verdict", recorded with its cause.
//
// Only --dry-run exists yet: the gate records the verdict and changes
// nothing on the reviewed issue.
//
// Exit codes: 0 for a valid verdict; 1 for no verdict, after its row is
// written, or for a GitHub failure; 2 for a usage or environment error.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/partio-io/cli/internal/review"
)

const usage = "usage: minion-review gate --issue <number> --verdict <path> --dry-run [--night YYYY-MM-DD]"

func main() {
	if len(os.Args) < 2 || os.Args[1] != "gate" {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	os.Exit(gate(os.Args[2:]))
}

func gate(args []string) int {
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	var (
		issue   = fs.Int("issue", 0, "reviewed issue number (required)")
		verdict = fs.String("verdict", "", "path to the verdict file (required)")
		dryRun  = fs.Bool("dry-run", false, "record the verdict only; required in this version")
		// The sweep passes one night per run, so a run that crosses
		// midnight still writes one comment.
		night = fs.String("night", time.Now().UTC().Format(time.DateOnly), "UTC date of the night comment")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *issue <= 0 || *verdict == "" || !*dryRun {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	if _, err := time.Parse(time.DateOnly, *night); err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: --night %q is not a YYYY-MM-DD date\n", *night)
		return 2
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	token := os.Getenv("GH_TOKEN")
	if repo == "" || token == "" {
		fmt.Fprintln(os.Stderr, "minion-review: GITHUB_REPOSITORY and GH_TOKEN must be set")
		return 2
	}
	api := os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}

	res, err := review.Run(review.Config{
		VerdictPath: *verdict,
		Repo:        repo,
		Issue:       *issue,
		Night:       *night,
		DryRun:      *dryRun,
		APIBaseURL:  api,
		Token:       token,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
		return 1
	}
	if !res.Valid {
		fmt.Println("minion-review: no verdict:", res.Cause)
		return 1
	}
	fmt.Println("minion-review:", res.Outcome, "(dry run)")
	return 0
}
