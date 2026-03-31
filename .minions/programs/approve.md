---
id: approve
target_repos:
  - cli
---

# Auto-approve eligible proposals

Scan open proposal issues and approve those that have passed the review window.

## Steps

1. **List open proposals:**
   ```bash
   gh issue list --repo <this-repo> --label minion-proposal --state open --json number,title,labels,createdAt --limit 200
   ```

2. **For each proposal**, check if it should be approved:
   - Skip if it has a `do-not-build` label
   - Skip if it already has `minion-approved` label
   - Skip if it was created less than 24 hours ago
   - Skip if it has `minion-executing` or `minion-failed` label

3. **Approve eligible proposals** by adding the `minion-approved` label:
   ```bash
   gh issue edit <number> --repo <this-repo> --add-label minion-approved
   ```

4. **Print summary** of approved and skipped proposals with reasons.
