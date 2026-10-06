package github

import (
	"fmt"
	"net/http"
	"strings"
)

// PerPage is GitHub's maximum page size for list endpoints.
const PerPage = 100

// GetPages walks the paginated list endpoint at path under c.BaseURL,
// PerPage items at a time, and hands each page to visit. The walk
// stops on a short page, when visit returns false, or after maxPages
// pages; maxPages zero means no bound.
func GetPages[T any](c Client, path string, maxPages int, visit func([]T) bool) error {
	url := c.BaseURL + path
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s%sper_page=%d&page=%d", url, sep, PerPage, page), nil)
		if err != nil {
			return err
		}
		var items []T
		if err := c.Do(req, &items); err != nil {
			return err
		}
		if !visit(items) || len(items) < PerPage {
			return nil
		}
	}
	return nil
}
