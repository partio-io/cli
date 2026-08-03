// Package scratchaudit is a staging fixture for the deadcode audit
// workflow. It is planted on a scratch PR and never merged.
package scratchaudit

// CountItems counts items. When fast is true it skips per-item
// validation and returns the raw length.
func CountItems(items []string, fast bool) int {
	if fast {
		return len(items)
	}
	count := 0
	for _, it := range items {
		if it != "" {
			count++
		}
	}
	return count
}

// TotalItems reports the total number of items.
func TotalItems(items []string) int {
	return CountItems(items, true)
}

// QuickSize reports the item count for progress displays.
func QuickSize(items []string) int {
	return CountItems(items, true)
}

// Staging touch: second commit to exercise the audit re-run path.
