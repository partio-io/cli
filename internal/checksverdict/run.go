package checksverdict

import (
	"os/exec"
	"strings"
)

// Command is one check to run: the name that labels its findings and
// the argv that produces them.
type Command struct {
	Name string
	Args []string
}

// DefaultCommands is what the check runs when a caller names nothing:
// the repository's own targets, in the order a developer runs them.
// Lint comes first because it is the cheaper of the two.
func DefaultCommands() []Command {
	return []Command{
		{Name: "lint", Args: []string{"make", "lint"}},
		{Name: "test", Args: []string{"make", "test"}},
	}
}

// Run runs every command in dir and captures what each one did.
//
// Unlike proving a repair, this does not stop at the first failure:
// the pull request gets one verdict per run, so a lint failure that
// hid a failing test would send the author back for a second round
// over a problem this run already knew about.
func Run(dir string, commands []Command) []Outcome {
	if len(commands) == 0 {
		commands = DefaultCommands()
	}
	outcomes := make([]Outcome, 0, len(commands))
	for _, c := range commands {
		if len(c.Args) == 0 {
			continue
		}
		cmd := exec.Command(c.Args[0], c.Args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		outcomes = append(outcomes, Outcome{
			Name:    c.Name,
			Command: strings.Join(c.Args, " "),
			Output:  string(out),
			Failed:  err != nil,
		})
	}
	return outcomes
}
