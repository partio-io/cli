package auditgate

import (
	"encoding/json"
	"fmt"
	"os"
)

// Finding is one audit finding: where the problem is and why the
// auditor believes it is real.
type Finding struct {
	Location  string `json:"location"`
	Reasoning string `json:"reasoning"`
}

// Verdict is the contract between an audit program and the gate:
// the program writes it as its final act, the gate maps it to a
// check outcome.
type Verdict struct {
	Status   string    `json:"status"`
	Findings []Finding `json:"findings"`
}

// loadVerdict reads and validates the verdict file. Any error —
// missing file, bad JSON, unknown status — is returned so the gate
// can fail closed.
func loadVerdict(path string) (Verdict, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Verdict{}, fmt.Errorf("read verdict: %w", err)
	}
	var v Verdict
	if err := json.Unmarshal(data, &v); err != nil {
		return Verdict{}, fmt.Errorf("parse verdict: %w", err)
	}
	if v.Status != "pass" && v.Status != "fail" {
		return Verdict{}, fmt.Errorf("verdict status %q is neither pass nor fail", v.Status)
	}
	if v.Status == "fail" && len(v.Findings) == 0 {
		return Verdict{}, fmt.Errorf("fail verdict carries no findings")
	}
	for i, f := range v.Findings {
		if f.Location == "" || f.Reasoning == "" {
			return Verdict{}, fmt.Errorf("finding %d is missing location or reasoning", i+1)
		}
	}
	return v, nil
}
