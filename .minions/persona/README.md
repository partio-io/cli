# persona substrate

This directory is the **only** decision-making substrate for the
`persona` agent in `.minions/programs/research.md`. The persona answers
research questions the way the maintainer would, and it grounds those
answers in the files under `telos/` and `memory/` here — nothing else.

## Why it lives here

The persona must be grounded **only** by sanitized, in-repo files — never
by cloning or reading a private or personal repository at run time. A
public CI pipeline that pulled private context and then posted to public
issues would be a private→public exposure path; keeping the substrate
in-repo and sanitized removes that path entirely.

## Privacy contract — READ BEFORE EDITING

This substrate lives in a **public** repository. Everything here is
world-readable and may be quoted by an LLM into public issue comments.
Therefore these files contain **no personal data**:

- No health / training / recovery data.
- No diary, journal, or personal-reflection content.
- No financial figures, budgets, or income.
- No calendar, schedule, location, or employer-internal detail.
- No names of private individuals.
- No content copied verbatim from any private source.

Only generic, publishable decision-making essence — engineering
trade-offs, partio product priorities, the quality bar, defer-domains,
and working preferences — belongs here.

## Refreshing this substrate (manual, local-only)

When the persona's guidance drifts from how the maintainer actually
decides, refresh it **manually and locally**:

1. A human curates these files from their own private notes, on their own
   machine.
2. They distill only publishable essence into the files here.
3. They review the diff against the privacy contract above for leaks.
4. They commit via a normal PR for review.

**Never** automate a clone or sync of any private or personal repository
into this (or any) public repo. Automation is the exposure path this
design exists to remove.
