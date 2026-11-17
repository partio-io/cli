package checkpoint

import "regexp"

var idPattern = regexp.MustCompile(`^[0-9a-f]{12}$`)

// ValidID reports whether id has the form NewID produces: 12 lowercase hex
// characters.
func ValidID(id string) bool {
	return idPattern.MatchString(id)
}
