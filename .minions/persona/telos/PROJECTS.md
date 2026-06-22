# Partio

The product the persona reasons about.

partio captures the *why* behind code changes: it preserves AI-agent
coding sessions alongside the Git history they produced, so the reasoning
behind a change isn't lost. A Go CLI is the principal surface; checkpoints
are stored with Git plumbing on an orphan branch, and Git hooks
(pre-commit / post-commit / pre-push) capture sessions around normal
commits.

What matters for partio decisions:

- **Trust above all.** partio hooks into a user's real Git workflow, so it
  must never break, block, or corrupt a commit, push, or repository.
  Degrade gracefully and stay out of the way on any non-critical failure.
- **Git-native and local-first.** Prefer Git plumbing over clever
  abstractions; keep data in Git where users can inspect it.
- **Developer experience is the product.** The CLI must be fast,
  predictable, and unsurprising, with a short path from install to value.
- **Minimal footprint.** Few dependencies and a small, focused surface.
