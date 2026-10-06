// Package github is the GitHub REST client the minion tools share.
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Client holds what every GitHub request needs.
type Client struct {
	BaseURL    string       // GitHub API root, e.g. https://api.github.com
	Token      string       // sent as a bearer token
	HTTPClient *http.Client // nil means http.DefaultClient
}

// StatusError is the error Do returns for a non-2xx response. Callers
// that must tell "absent" from "broken" check Code with errors.As.
type StatusError struct {
	Method string
	Path   string
	Code   int
	Status string
	Detail []byte // up to 512 bytes of the response body
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("github: %s %s: %s: %s", e.Method, e.Path, e.Status, e.Detail)
}

// Do executes one GitHub API request and decodes the response into out
// when out is non-nil. A non-2xx response is a *StatusError that names
// the method, the path, the status and up to 512 bytes of the body.
func (c Client) Do(req *http.Request, out any) error {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		// Drain what is left so the connection can be reused.
		if _, drainErr := io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10)); drainErr != nil {
			slog.Debug("draining response body", "url", req.URL.Path, "error", drainErr)
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Debug("closing response body", "url", req.URL.Path, "error", closeErr)
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, readErr := io.ReadAll(io.LimitReader(resp.Body, 512))
		if readErr != nil {
			detail = fmt.Appendf(nil, "(error body unreadable: %v)", readErr)
		}
		return &StatusError{Method: req.Method, Path: req.URL.Path, Code: resp.StatusCode, Status: resp.Status, Detail: detail}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
