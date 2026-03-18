package hooks

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	beginSentinel = "# BEGIN partio"
	endSentinel   = "# END partio"
)

var hookNames = []string{"pre-commit", "post-commit", "pre-push"}

var partioBlockRe = regexp.MustCompile(`\n?` + beginSentinel + `\n[\s\S]*?` + endSentinel + `\n?`)

// partioBlock returns the partio invocation block for the given hook name.
func partioBlock(name string) string {
	return fmt.Sprintf(`%s
if command -v partio &> /dev/null; then
    partio _hook %s "$@"
    exit_code=$?
    [ $exit_code -ne 0 ] && exit $exit_code
fi
%s`, beginSentinel, name, endSentinel)
}

// newHookScript returns a complete hook script for a new hook file.
func newHookScript(name string) string {
	return "#!/bin/bash\n" + partioBlock(name) + "\n"
}

func hasPartioBlock(content string) bool {
	return strings.Contains(content, beginSentinel)
}

func removePartioBlock(content string) string {
	return partioBlockRe.ReplaceAllString(content, "")
}
