package patchapply

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestPushRemoteUsesThePersonalAccessToken pins which credential the
// round pushes with.
//
// A push made with the default Actions token triggers no follow-up
// workflow run, and the next repair round is a follow-up run. The fix
// would land, the loop would stop, and nothing would say so.
func TestPushRemoteUsesThePersonalAccessToken(t *testing.T) {
	cfg := Config{Repo: "partio-io/cli", Token: "ghp_example"}

	want := "https://x-access-token:ghp_example@github.com/partio-io/cli.git"
	if got := pushRemote(cfg); got != want {
		t.Errorf("pushRemote = %q, want %q", got, want)
	}

	cfg.Remote = "/srv/origin.git"
	if got := pushRemote(cfg); got != cfg.Remote {
		t.Errorf("pushRemote = %q, want the configured remote %q", got, cfg.Remote)
	}
}

// TestRunKeepsTheTokenOutOfWhatItReports drives a real failure whose
// output echoes the remote, because that is how a token reaches a job
// log: git prints the URL it could not read from.
func TestRunKeepsTheTokenOutOfWhatItReports(t *testing.T) {
	f := newFixture(t, false)
	patch := f.makePatch(t, "notes.txt", "one\ntwo\n")

	cfg := f.config(patch)
	cfg.Token = "ghp_example"
	cfg.Remote = filepath.Join(t.TempDir(), "ghp_example-missing.git")

	res, err := Run(cfg)
	if err == nil {
		t.Fatalf("Run succeeded against a remote that does not exist: %+v", res)
	}
	if strings.Contains(err.Error(), "ghp_example") {
		t.Errorf("the error carries the token: %v", err)
	}
	if !strings.Contains(err.Error(), "***") {
		t.Errorf("error = %v, want the token replaced", err)
	}
}
