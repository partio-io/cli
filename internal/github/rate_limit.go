package github

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

// withRateLimitHint adds a hint to set GH_TOKEN when err is a GitHub rate
// limit response: a 429, or a 403 whose body names the rate limit (GitHub
// also answers 403 for plain permission failures, which get no hint).
func withRateLimitHint(err error) error {
	var se *StatusError
	if !errors.As(err, &se) {
		return err
	}
	limited := se.Code == http.StatusTooManyRequests ||
		(se.Code == http.StatusForbidden && bytes.Contains(bytes.ToLower(se.Detail), []byte("rate limit")))
	if !limited {
		return err
	}
	return fmt.Errorf("%w (GitHub API rate limit reached; set GH_TOKEN to a GitHub token to raise the limit)", err)
}
