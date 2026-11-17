package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// syncForkRefPrefix is where fetched fork branches are parked. It sits
// outside refs/heads/ so the data never shows up as a local branch.
const syncForkRefPrefix = "refs/partio/sync-fork/"

// fetchForkBranch fetches the checkpoint branch of the repository at
// cloneURL into a temporary ref unique to this run. The fork is only ever
// read. The returned cleanup deletes the ref and must be called on success
// and on error alike; it is non-nil even when the fetch fails.
func fetchForkBranch(dir, cloneURL string) (ref string, cleanup func(), err error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", func() {}, fmt.Errorf("generating temporary ref name: %w", err)
	}
	ref = syncForkRefPrefix + hex.EncodeToString(b)
	cleanup = func() {
		del := exec.Command("git", "update-ref", "-d", ref)
		del.Dir = dir
		if out, delErr := del.CombinedOutput(); delErr != nil {
			slog.Warn("could not delete temporary ref", "ref", ref, "error", delErr, "output", strings.TrimSpace(string(out)))
		}
	}

	refspec := "+refs/heads/" + git.CheckpointBranch + ":" + ref
	// "--" keeps a hostile URL from being read as an option.
	cmd := exec.Command("git", "fetch", "--no-tags", "--no-write-fetch-head", "--", cloneURL, refspec)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return ref, cleanup, fmt.Errorf("fetching %s from %s: %w: %s", git.CheckpointBranch, cloneURL, err, strings.TrimSpace(string(out)))
	}
	return ref, cleanup, nil
}
