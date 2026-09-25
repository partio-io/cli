package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestFetchBranch(t *testing.T) {
	// initRepos creates a bare repo (remote) and a working repo pointing at it
	// as "origin", returning the two temp directories.
	initRepos := func(t *testing.T) (bareDir, workDir string) {
		t.Helper()
		bareDir = t.TempDir()
		workDir = t.TempDir()

		runIn := func(dir string, args ...string) {
			t.Helper()
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command %v in %s failed: %v\n%s", args, dir, err, out)
			}
		}

		runIn(bareDir, "git", "init", "--bare")
		runIn(workDir, "git", "init")
		runIn(workDir, "git", "config", "user.email", "test@example.com")
		runIn(workDir, "git", "config", "user.name", "Test")
		runIn(workDir, "git", "remote", "add", "origin", bareDir)
		return
	}

	// pushToBare pushes a single empty commit to the given branch in bareDir.
	pushToBare := func(t *testing.T, bareDir, branch string) {
		t.Helper()
		tmp := t.TempDir()
		runIn := func(args ...string) {
			t.Helper()
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = tmp
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command %v failed: %v\n%s", args, err, out)
			}
		}
		runIn("git", "init")
		runIn("git", "config", "user.email", "test@example.com")
		runIn("git", "config", "user.name", "Test")
		runIn("git", "commit", "--allow-empty", "-m", "checkpoint")
		runIn("git", "remote", "add", "origin", bareDir)
		runIn("git", "push", "origin", "HEAD:refs/heads/"+branch)
	}

	tests := []struct {
		name         string
		remoteArg    string // non-empty overrides the remote ("origin") used in the call
		branchArg    string // non-empty overrides CheckpointBranch used in the call
		extraSetup   func(t *testing.T, bareDir string)
		wantErr      bool
		checkAfter   func(t *testing.T, workDir string)
	}{
		{
			name: "branch exists on remote: updates tracking ref",
			extraSetup: func(t *testing.T, bareDir string) {
				pushToBare(t, bareDir, CheckpointBranch)
			},
			checkAfter: func(t *testing.T, workDir string) {
				t.Helper()
				cmd := exec.Command("git", "rev-parse",
					"refs/remotes/origin/"+CheckpointBranch)
				cmd.Dir = workDir
				out, err := cmd.Output()
				if err != nil {
					t.Errorf("refs/remotes/origin/%s should exist after fetch: %v",
						CheckpointBranch, err)
				}
				if strings.TrimSpace(string(out)) == "" {
					t.Errorf("refs/remotes/origin/%s should be non-empty after fetch",
						CheckpointBranch)
				}
			},
		},
		{
			name:    "branch absent from remote: returns nil without error",
			wantErr: false,
			checkAfter: func(t *testing.T, workDir string) {
				t.Helper()
				cmd := exec.Command("git", "rev-parse",
					"refs/remotes/origin/"+CheckpointBranch)
				cmd.Dir = workDir
				if err := cmd.Run(); err == nil {
					t.Errorf("refs/remotes/origin/%s should not exist when remote branch is absent",
						CheckpointBranch)
				}
			},
		},
		{
			name:      "unreachable remote: returns non-nil error",
			remoteArg: "/nonexistent/path/that/does/not/exist",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bareDir, workDir := initRepos(t)
			if tt.extraSetup != nil {
				tt.extraSetup(t, bareDir)
			}
			t.Chdir(workDir)

			remote := "origin"
			if tt.remoteArg != "" {
				remote = tt.remoteArg
			}
			branch := CheckpointBranch
			if tt.branchArg != "" {
				branch = tt.branchArg
			}

			err := FetchBranch(remote, branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkAfter != nil {
				tt.checkAfter(t, workDir)
			}
		})
	}
}
