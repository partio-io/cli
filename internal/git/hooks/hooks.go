package hooks

import (
	"fmt"
	"strings"
)

const partioMarker = "# Installed by partio"

var hookNames = []string{"pre-commit", "post-commit", "pre-push"}

// hookScript returns the bash shim for a given hook name.
func hookScript(name string) string {
	return fmt.Sprintf(`#!/bin/bash
%s
if command -v partio &> /dev/null; then
    partio _hook %s "$@"
    exit_code=$?
    [ $exit_code -ne 0 ] && exit $exit_code
fi
# Chain to original hook if backed up
hooks_dir="$(git rev-parse --git-common-dir)/hooks"
[ -f "$hooks_dir/%s.partio-backup" ] && exec "$hooks_dir/%s.partio-backup" "$@"
exit 0
`, partioMarker, name, name, name)
}

func isPartioHook(content string) bool {
	return strings.Contains(content, partioMarker)
}
