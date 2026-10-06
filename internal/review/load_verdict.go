package review

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadVerdict reads the verdict file at path for issue. Any error is
// "no verdict": the gate fails closed and reports the error as the
// cause.
func LoadVerdict(path string, issue int) (Verdict, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Verdict{}, fmt.Errorf("read verdict: %w", err)
	}
	var v Verdict
	if err := json.Unmarshal(data, &v); err != nil {
		return Verdict{}, fmt.Errorf("malformed verdict: %w", err)
	}
	if err := v.check(issue); err != nil {
		return Verdict{}, err
	}
	return v, nil
}
