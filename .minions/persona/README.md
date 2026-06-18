# persona substrate

This directory is the **only** decision-making substrate for the
`persona` agent in `.minions/programs/research.md`. The persona answers
research questions "the way the maintainer (jcleira) would," and it
grounds those answers in the files under `telos/` and `memory/` here —
nothing else.

## Why it lives here (and not in a private repo)

The persona used to be grounded by cloning a **private** personal repo
(`jcleira/argos`) into this **public** repo at CI time, then posting
PRD/slice comments on public issues. That was a private→public exposure
path. It has been removed (see
`docs/2026-05-05-research-minion/issues/08-persona-local-substrate.md`
in the minions docs, and partio-io/cli#454). The pipeline must **never**
clone or read argos, or any other private repo.

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
trade-offs, product priorities for the public products, the quality bar,
defer-domains, and working preferences — belongs here.

## Refreshing this substrate (manual, local-only)

When the persona's guidance drifts from how the maintainer actually
decides, refresh it **manually and locally**:

1. A human reads the private source (argos) on their own machine.
2. They hand-edit the files here, distilling only publishable essence.
3. They review the diff against the privacy contract above for leaks.
4. They commit via a normal PR for review.

**Never** automate a clone or sync of argos (or any private repo) into
this or any public repo. Automation is the exposure path this design
exists to remove.
