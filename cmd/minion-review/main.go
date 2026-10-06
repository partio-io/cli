// Command minion-review is the deterministic half of the proposal
// review. Its gate subcommand reads the verdict file the review program
// wrote for one issue, checks it, acts on the issue, and records the
// verdict as one row in the tracking issue. A keep gets an evidence
// comment and the minion-reviewed label; a close gets the comment, the
// label and the close. A rewrite is no verdict yet. It fails closed: a
// missing or invalid verdict is "no verdict", recorded with its cause.
//
// With --dry-run the gate records the verdict and changes nothing on
// the reviewed issue. With --mode build it runs as the first step of a
// build: it acts, marks its row as a build, and appends the blocked and
// changed step outputs to the file GITHUB_OUTPUT names.
//
// Its dupes subcommand lists the minion-proposal issues that may hold
// the same idea as one issue, by source item or by title, and prints
// them as JSON. The review program calls it; it changes nothing.
//
// Its next subcommand prints the issues the sweep reviews next, one
// number per line: the open minion-proposal issues without
// minion-reviewed, approved ones first, then oldest first. It skips an
// issue the tracking issue lists under "needs you". With --sample N it
// prints N of them, spread across the filing months, for a dry run. It
// changes nothing.
//
// Exit codes: 0 for a valid verdict, a dupes list or a batch; 1 for
// no verdict, after its row is written, or for a GitHub failure; 2 for
// a usage or environment error.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/partio-io/cli/internal/github"
	"github.com/partio-io/cli/internal/review"
)

const (
	gateUsage  = "usage: minion-review gate --issue <number> --verdict <path> [--mode sweep|build] [--dry-run] [--night YYYY-MM-DD]"
	dupesUsage = "usage: minion-review dupes --title <title> [--source <ref>]... [--exclude <number>]"
	nextUsage  = "usage: minion-review next [--sample <n>]"
	usage      = gateUsage + "\n" + dupesUsage + "\n" + nextUsage
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "gate":
		os.Exit(gate(os.Args[2:]))
	case "dupes":
		os.Exit(dupes(os.Args[2:], os.Stdout))
	case "next":
		os.Exit(next(os.Args[2:], os.Stdout))
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
}

// githubEnv reads the repo, token and API root the workflow sets. It
// returns false, after it prints the cause, when one is missing.
func githubEnv() (repo, token, api string, ok bool) {
	repo = os.Getenv("GITHUB_REPOSITORY")
	token = os.Getenv("GH_TOKEN")
	if repo == "" || token == "" {
		fmt.Fprintln(os.Stderr, "minion-review: GITHUB_REPOSITORY and GH_TOKEN must be set")
		return "", "", "", false
	}
	api = os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}
	return repo, token, api, true
}

func gate(args []string) int {
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	var (
		issue   = fs.Int("issue", 0, "reviewed issue number (required)")
		verdict = fs.String("verdict", "", "path to the verdict file (required)")
		dryRun  = fs.Bool("dry-run", false, "record the verdict only; change nothing on the reviewed issue")
		mode    = fs.String("mode", "sweep", "sweep, or build for the review a build runs first")
		// The sweep passes one night per run, so a run that crosses
		// midnight still writes one comment.
		night = fs.String("night", time.Now().UTC().Format(time.DateOnly), "UTC date of the night comment")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *issue <= 0 || *verdict == "" || fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, gateUsage)
		return 2
	}
	build := *mode == "build"
	if (*mode != "sweep" && !build) || (build && *dryRun) {
		fmt.Fprintln(os.Stderr, gateUsage)
		return 2
	}
	output := os.Getenv("GITHUB_OUTPUT")
	if build && output == "" {
		fmt.Fprintln(os.Stderr, "minion-review: --mode build needs GITHUB_OUTPUT")
		return 2
	}
	if _, err := time.Parse(time.DateOnly, *night); err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: --night %q is not a YYYY-MM-DD date\n", *night)
		return 2
	}
	repo, token, api, ok := githubEnv()
	if !ok {
		return 2
	}

	res, err := review.Run(review.Config{
		VerdictPath: *verdict,
		Repo:        repo,
		Issue:       *issue,
		Night:       *night,
		DryRun:      *dryRun,
		Build:       build,
		APIBaseURL:  api,
		Token:       token,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
		// The gate can fail after it acted, on the tracking row. A
		// build still gets blocked=true after a close it made, so the
		// failure step does not mark a closed issue as a failed build.
		if build && res.Valid {
			if werr := writeOutputs(output, res); werr != nil {
				fmt.Fprintf(os.Stderr, "minion-review: %v\n", werr)
			}
		}
		return 1
	}
	if !res.Valid {
		fmt.Println("minion-review: no verdict:", res.Cause)
		return 1
	}
	if build {
		if err := writeOutputs(output, res); err != nil {
			fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
			return 1
		}
	}
	suffix := ""
	switch {
	case *dryRun:
		suffix = " (dry run)"
	case build:
		suffix = " (build)"
	}
	fmt.Println("minion-review:", res.Outcome+suffix)
	return 0
}

// writeOutputs appends the build's step outputs to the GITHUB_OUTPUT
// file.
func writeOutputs(path string, res review.Result) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	err = review.WriteBuildOutputs(f, res)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("write GITHUB_OUTPUT: %w", err)
	}
	return nil
}

// sourceFlags collects each --source value.
type sourceFlags []string

func (s *sourceFlags) String() string     { return strings.Join(*s, " ") }
func (s *sourceFlags) Set(v string) error { *s = append(*s, v); return nil }

func dupes(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("dupes", flag.ContinueOnError)
	var sources sourceFlags
	fs.Var(&sources, "source", "a source reference of the issue, in any form: owner/repo#N, a URL, an Origin: line (repeatable)")
	var (
		title   = fs.String("title", "", "title of the issue (required)")
		exclude = fs.Int("exclude", 0, "the issue under review, left out of the candidates")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*title) == "" || fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, dupesUsage)
		return 2
	}
	for _, src := range sources {
		if len(review.SourceRefs(src)) == 0 {
			fmt.Fprintf(os.Stderr, "minion-review: --source %q names no owner/repo#N item; it matches nothing\n", src)
		}
	}
	repo, token, api, ok := githubEnv()
	if !ok {
		return 2
	}

	found, err := review.FindDupes(github.Client{BaseURL: api, Token: token}, review.DupesQuery{
		Repo: repo, Sources: sources, Title: *title, Exclude: *exclude,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
		return 1
	}
	if found == nil {
		found = []review.Dupe{}
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(found); err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
		return 1
	}
	return 0
}

func next(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("next", flag.ContinueOnError)
	sample := fs.Int("sample", 0, "pick n issues spread across the filing months, for a dry run")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	sampled := false
	fs.Visit(func(f *flag.Flag) { sampled = sampled || f.Name == "sample" })
	if fs.NArg() > 0 || *sample < 0 || (sampled && *sample == 0) {
		fmt.Fprintln(os.Stderr, nextUsage)
		return 2
	}
	repo, token, api, ok := githubEnv()
	if !ok {
		return 2
	}

	numbers, err := review.Next(github.Client{BaseURL: api, Token: token}, review.NextQuery{
		Repo: repo, Sample: *sample,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
		return 1
	}
	for _, n := range numbers {
		if _, err := fmt.Fprintln(stdout, n); err != nil {
			fmt.Fprintf(os.Stderr, "minion-review: %v\n", err)
			return 1
		}
	}
	return 0
}
