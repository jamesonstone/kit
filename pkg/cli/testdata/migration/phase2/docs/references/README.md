# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature; keep long-lived background here instead of in agent entry files
- `rules/<slug>.md` holds contextual rules; the universal contract's Contextual Rules section says when to read each one, and `kit rules list` is the live inventory
- `workflows/<slug>.md` holds declarative workflow evidence lists used by `kit context resolve`
- `worktrees.md` describes the native Git worktree hierarchy, naming, shared-state model, and environment links
- Link these files from feature front matter references when they materially shape work

## Rules Management

- Use `kit rules add` to import or activate registry rulesets and `kit rules view <slug>` to preview one
- Use `kit init --refresh` to adopt registry rules into `.kit.yaml` state and pick up safe upstream updates
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder

## Starter Files

- `testing.md` — the project's validation commands, suites, and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations
