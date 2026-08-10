package repairround

import (
	"fmt"
	"strings"
)

// markerPrefix opens a repair commit's subject. What follows, up to
// the closing parenthesis and ": ", names the check that produced the
// commit.
const markerPrefix = "audit("

// markerInfix closes the check name and opens the human-readable rest
// of the subject.
const markerInfix = "): "

// Subject returns the commit subject a repair round must use. The
// check name is part of the subject so that a later run can tell which
// check spent the round by reading the branch's subjects alone.
func Subject(check string, pr int) string {
	return fmt.Sprintf("%s%s%sfix findings (#%d)", markerPrefix, check, markerInfix, pr)
}

// CheckOf reports the check named by a repair commit's subject.
//
// The match is deliberately strict, because miscounting here spends a
// budget that belongs to someone else. The marker must open the
// subject, the check name must be lowercase kebab-case, and the
// closing "): " must be present. A subject that merely mentions a
// marker, an older unqualified "audit:" subject, or a human commit
// with a similar shape all report false.
func CheckOf(subject string) (string, bool) {
	rest, ok := strings.CutPrefix(subject, markerPrefix)
	if !ok {
		return "", false
	}
	check, _, ok := strings.Cut(rest, markerInfix)
	if !ok || !validCheck(check) {
		return "", false
	}
	return check, true
}

// validCheck reports whether name is a well-formed check name:
// non-empty lowercase letters, digits and hyphens.
func validCheck(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return true
}
