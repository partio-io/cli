---
id: implement
target_repos:
  - cli
acceptance_criteria:
  - "All changes match what the issue describes"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Implement the issue

Read the issue provided as context and implement exactly what it describes.

Follow existing code patterns and conventions. Read the relevant code before making changes. Keep changes minimal — implement what the issue asks for, nothing more.

After implementation, run the appropriate checks (`make test`, `make lint`) and fix any failures.
