# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- Keep long-lived background context here instead of in injected top-level instruction files
- Link these files from feature front matter references when they materially shape work
- Store durable rulesets under `rules/<slug>.md` and link them with `kit rules link` instead of copying rules into agent instruction files
- Use `kit rules add` to import or activate available registry rulesets from the Kit GitHub `main` branch
- Use `kit rules view <slug>` to preview a local or registry ruleset before importing it
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder
- `kit rule` is the singular alias for `kit rules`

## Starter Files

- `testing.md` — repo-wide testing norms and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations
- `rules/` — pointer-loaded durable rulesets such as frontend UI rules, testing rules, API conventions, security constraints, or domain rules
