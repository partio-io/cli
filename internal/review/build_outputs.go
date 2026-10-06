package review

import (
	"fmt"
	"io"
)

// WriteBuildOutputs writes the step outputs a build reads after its
// review, in the GITHUB_OUTPUT format: blocked is true after a close,
// so the build stops; changed is true after a rewrite, so the old
// research plan no longer counts. A close writes no changed, because
// nothing runs after it. A result that is not Valid writes nothing: no
// verdict fails the job instead.
func WriteBuildOutputs(w io.Writer, res Result) error {
	if !res.Valid {
		return nil
	}
	out := "blocked=true\n"
	if res.Outcome != OutcomeClose {
		out = fmt.Sprintf("blocked=false\nchanged=%t\n", res.Outcome == OutcomeRewrite)
	}
	_, err := io.WriteString(w, out)
	return err
}
