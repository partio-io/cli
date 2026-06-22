# Engineering principles

- **Test behavior through the public interface**, not implementation
  details. A good test survives an internal refactor.
- **Table-driven tests, standard library only.** No external test
  frameworks; isolate the filesystem and environment per test.
- **Minimal dependencies.** Every dependency is a liability; add one only
  when it clearly earns its place.
- **One primary concern per file**, with a descriptive name.
- **DDD-influenced structure.** Domain types live alongside their logic in
  focused packages, with clear boundaries between domain, storage, and
  transport.
- **Error resilience at the edges.** Code that hooks into a user's Git
  workflow logs and degrades gracefully on non-critical failures — it
  never blocks a commit or push.
- **Match the conventions already in the repository** being changed rather
  than importing new ones.
