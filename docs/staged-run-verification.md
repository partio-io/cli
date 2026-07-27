# Staged-run verification

## Overview

A staged minion run verifies that a slice-aware build correctly produces `minion:slice N/M` marker commits on a single branch across multiple independent sessions. Each slice session builds its assigned slice and commits with the appropriate marker, ensuring all work lands on one branch with a single PR and closes through the normal done flow.

## Checklist

After a staged run, eyeball the following:

- [ ] `minion:slice N/M` marker commits are present for each slice (e.g. `minion:slice 1/2`, `minion:slice 2/2`)
- [ ] All slice work is on a single `minion/implement-implement-<n>` branch with one PR
- [ ] The issue was closed by the normal done flow (not manually)
