---
id: configurable-review-roles
target_repos:
  - cli
acceptance_criteria:
  - Users can define review roles (e.g., security, performance, correctness) in Partio settings
  - Each role specifies a name, description, and focus areas that guide the review
  - partio review accepts a --role flag to select a configured role
  - Review output includes the role context alongside checkpoint-enriched review findings
  - Default roles are provided out of the box (general, security, performance)
pr_labels:
  - minion
---

# Add configurable review roles and tuning settings for code review

Extend the review capability with user-configurable roles and tuning settings. Roles define the perspective and focus areas for AI-assisted code review, allowing teams to run targeted reviews (security audit, performance review, correctness check) that leverage checkpoint context to understand the reasoning behind changes.

## What to implement

1. Add a `review_roles` section to Partio settings schema where users can define named review roles with descriptions and focus areas.
2. Provide sensible default roles: `general` (broad review), `security` (focus on vulnerabilities, auth, input validation), `performance` (focus on hot paths, allocations, complexity).
3. Add a `--role` flag to the review command that selects a configured role.
4. Add `review_settings` to configuration for tuning parameters like review depth, context window size, and whether to include full checkpoint transcripts.
5. When a role is selected, include its description and focus areas in the review context alongside the checkpoint data.

## Why this matters

Different review scenarios need different perspectives. A security review before a release needs to focus on different things than a daily code quality check. Configurable roles let teams codify their review standards and run consistent, targeted reviews that use checkpoint context to understand not just what changed but why — enabling more informed review feedback.
