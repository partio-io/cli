// Command minion-repair-round reports whether a pull request may
// receive another automated repair round for one check, and which
// round it would be. It replaces the loop guard that skipped any run
// whose head commit looked like a repair: a repair that lands one edit
// short of green now gets another attempt, up to the cap.
//
// It writes skip, round, prior and subject to $GITHUB_OUTPUT. With
// --repairable it also answers the question that comes before the
// count — may this pull request receive a repair commit at all — and
// writes repairable and reason. A gate that runs on every pull request
// needs that answer; a gate that already guards on the label does not.
//
// Exit 1 means the count could not be established — the caller must
// not audit on a number it does not have.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/partio-io/cli/internal/repairround"
)

func main() {
	var (
		pr         = flag.Int("pr", 0, "pull request number (required)")
		check      = flag.String("check", "", "check whose budget is counted, e.g. dead-code (required)")
		maxRounds  = flag.Int("max", repairround.DefaultMaxRounds, "repair rounds allowed for this check")
		repairable = flag.Bool("repairable", false,
			"also decide whether this pull request may be repaired at all, and report it")
	)
	flag.Parse()
	if *pr == 0 || *check == "" {
		fmt.Fprintln(os.Stderr,
			"usage: minion-repair-round --pr <number> --check <name> [--max <n>] [--repairable]")
		os.Exit(2)
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	token := os.Getenv("GH_TOKEN")
	if repo == "" || token == "" {
		fmt.Fprintln(os.Stderr, "minion-repair-round: GITHUB_REPOSITORY and GH_TOKEN must be set")
		os.Exit(2)
	}
	api := os.Getenv("GITHUB_API_URL")
	if api == "" {
		api = "https://api.github.com"
	}

	cfg := repairround.Config{
		Repo:       repo,
		PRNumber:   *pr,
		Check:      *check,
		MaxRounds:  *maxRounds,
		APIBaseURL: api,
		Token:      token,
	}

	// Asked for and reported before the count, because the two answers
	// are different: an ineligible pull request was never the
	// pipeline's to touch, while a spent budget is one it tried.
	var eligibility repairround.Eligibility
	if *repairable {
		e, err := repairround.RunEligibility(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "minion-repair-round: %v\n", err)
			os.Exit(1)
		}
		eligibility = e
		if eligibility.Allowed {
			fmt.Printf("minion-repair-round: %s may repair this pull request — %s\n", *check, eligibility.Reason)
		} else {
			fmt.Printf("minion-repair-round: %s will not repair this pull request — %s\n", *check, eligibility.Reason)
		}
	}

	// Counted even for a pull request no repair may touch, so the round
	// number the gate reports stays true if a label is removed part-way
	// through a repair loop.
	d, err := repairround.Run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "minion-repair-round: %v\n", err)
		os.Exit(1)
	}

	if d.Allowed {
		fmt.Printf("minion-repair-round: %s round %d of %d may start (%d already on the branch)\n",
			*check, d.Round, *maxRounds, d.Prior)
	} else {
		fmt.Printf("minion-repair-round: %s spent its repair budget (%d of %d rounds); not auditing again\n",
			*check, d.Prior, *maxRounds)
	}

	if err := writeOutputs(d, eligibility, *repairable, *check, *pr); err != nil {
		fmt.Fprintf(os.Stderr, "minion-repair-round: %v\n", err)
		os.Exit(1)
	}
}

// writeOutputs appends the decision to the step's $GITHUB_OUTPUT file.
// The subject travels with it so the commit a repair pushes is built
// by the same package that later counts it.
//
// repairable and reason are written only when they were asked for. A
// caller that did not ask gets no such outputs, so a workflow cannot
// read a "false" that was never decided.
func writeOutputs(d repairround.Decision, e repairround.Eligibility, asked bool, check string, pr int) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	out := fmt.Sprintf("skip=%t\nround=%d\nprior=%d\nsubject=%s\n",
		!d.Allowed, d.Round, d.Prior, repairround.Subject(check, pr))
	if asked {
		// The reason is one line by construction; a multi-line value
		// would need a heredoc delimiter and there is nothing to gain
		// from allowing one.
		out += fmt.Sprintf("repairable=%t\nreason=%s\n", e.Allowed, singleLine(e.Reason))
	}
	_, writeErr := fmt.Fprint(f, out)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("write GITHUB_OUTPUT: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close GITHUB_OUTPUT: %w", closeErr)
	}
	return nil
}

// singleLine folds a reason onto one line so it cannot break the
// key=value form $GITHUB_OUTPUT expects.
func singleLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
