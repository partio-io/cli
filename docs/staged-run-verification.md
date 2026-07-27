# Staged-run verification

## Overview

A staged minion run verifies that a slice-aware build correctly produces `minion:slice N/M` marker commits on a single branch across multiple independent sessions. Each slice session builds its assigned slice and commits with the appropriate marker, ensuring all work lands on one branch with a single PR and closes through the normal done flow.
