# Staged-run verification

## Overview

A staged minion run verifies that a slice-aware build produces the expected `minion:slice N/M` marker commits on a single `minion/implement-*` branch throughout the multi-slice execution. It confirms that each slice session completes independently, that all work lands on one branch with a single PR, and that the issue is closed by the normal done flow after all slices finish.

## Checklist

After a staged run, eyeball the following:

- [ ] `minion:slice N/M` marker commits are present on the branch (one per slice)
- [ ] All slice work landed on a single `minion/implement-<issue>` branch with one PR
- [ ] The issue was closed by the normal done flow (not manually)
