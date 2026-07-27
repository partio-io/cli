# Staged-run verification

## Overview

A staged minion run verifies that a slice-aware build correctly produces `minion:slice N/M` marker commits on a single branch across multiple independent sessions. Each slice session builds its assigned slice and commits with the appropriate marker, ensuring all work lands on one branch with a single PR and closes through the normal done flow.

## Checklist

After a staged run, eyeball the following:

- [ ] `minion:slice N/M` marker commits are present — one per slice, in order, on the branch.
- [ ] All work is on a single minion branch with one PR (no extra branches or duplicate PRs).
- [ ] The issue is closed by the normal done flow (not manually).
