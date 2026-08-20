The premise of issue #30, filed 2026-03-26 and built on 2026-08-10.

Both claims are false for this repository. They came from a sibling
product's changelog, where the problem was real, and nobody checked
whether it was real here. The hook runs a two-commit `git diff`, which
is a tree-to-tree comparison and never touches the working tree.

Each line names the evidence that settles it. This file only captures
the premise; verifying it is slice 4.

## Premise

<!-- partio:premise:v1 -->

- Partio's post-commit hook performs a full tree walk to determine which files were changed in a commit. [evidence: `internal/hooks/postcommit.go`]
- The cost of determining changed files is O(N) in the number of tracked files. [evidence: `internal/attribution/calculate.go`]
