# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- Keep long-lived background context here instead of in injected top-level instruction files
- Link these files from feature front matter references when they materially shape work
- Store durable rulesets under `rules/<slug>.md` and link them with `kit rules link` instead of copying rules into agent instruction files
- Use `rules/kit-capabilities-usage.md` in downstream projects for Kit command discovery guidance
- Use `rules/command-capabilities.md` only inside the Kit source repository for maintainer command-surface work
- Use `rules/agent-team-orchestration.md` when deciding supervisor/subagent topology, concurrency, file-overlap handling, or read-only verification lanes
- Use `rules/feature-notes.md` when deciding how to load, reference, promote, or ignore source material under `docs/notes/<feature>`
- Use `rules/readme-header-tagline.md` when creating or materially updating top-level README openings
- Use `kit rules add` to import or activate available registry rulesets from the Kit GitHub `main` branch
- Use `kit rules view <slug>` to preview a local or registry ruleset before importing it
- Use `kit reconcile --include-files` to adopt existing registry rules into `.kit.yaml` registry state and pick up safe upstream ruleset updates
- Use `kit reconcile --include-files --force --file=docs/references/rules/<slug>.md` only when you intentionally want to accept the latest registry content for a customized ruleset
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder
- `kit rule` is the singular alias for `kit rules`

## Starter Files

- `testing.md` — repo-wide testing norms and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations
- `rules/` — pointer-loaded durable rulesets such as frontend UI rules, testing rules, API conventions, security constraints, or domain rules
- `../notes/<feature>/` — optional feature source material; not canonical truth and private contents remain ignored
