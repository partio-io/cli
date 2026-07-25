# Staged-run verification

## Overview

A staged minion run verifies that a slice-aware build produces the expected `minion:slice N/M` marker commits on a single `minion/implement-*` branch throughout the multi-slice execution. It confirms that each slice session completes independently, that all work lands on one branch with a single PR, and that the issue is closed by the normal done flow after all slices finish.
