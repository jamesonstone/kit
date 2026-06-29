---
kind: ruleset
slug: feature-notes
description: Guides agents to use docs/notes/<feature> as optional source material while keeping canonical truth in SPEC.md and durable project docs.
status: active
registry_scope: downstream
applies_to:
  - notes
  - feature-notes
  - source-material
  - context
  - documentation
  - coding-agent
read_policy_default: conditional
---

# Ruleset: feature-notes

## Purpose

- Make feature notes useful as source material without making them canonical truth.
- Preserve raw context, customer or Slack excerpts, screenshots, research, draft responses, and local-only notes in predictable locations.
- Keep durable requirements, decisions, constraints, and evidence in `SPEC.md`, `docs/CONSTITUTION.md`, or durable references.

## Applies When

- The task asks to use notes, feature notes, Slack history, customer context, research material, screenshots, draft responses, or source material.
- A feature front matter reference points at `docs/notes/<feature>` or a file under that directory.
- `kit notes`, `kit spec`, `kit map`, `kit summarize`, `kit handoff`, or an agent prompt surfaces a feature notes path.
- The current decision depends on prior raw context that may live outside `SPEC.md`.

## Structure

Use the standard feature notes scaffold:

- `docs/notes/<feature>/README.md` — directory contract and feature pointer.
- `docs/notes/<feature>/inbox/` — raw captured inputs, conversation excerpts, and transient context.
- `docs/notes/<feature>/references/` — source material, links, screenshots, research, and external references.
- `docs/notes/<feature>/responses/` — draft or sent responses tied to the feature.
- `docs/notes/<feature>/private/` — local-only ignored context; only `README.md` and `.gitignore` are tracked.

## Rules

- Treat notes as source material, not authoritative requirements.
- Load notes through RLM: list or inspect the smallest relevant set before reading note contents.
- Do not read every note file by default.
- Ignore `.gitkeep` files and empty placeholders.
- Read `README.md` first when the note directory contract or feature pointer is unclear.
- Read `inbox/`, `references/`, and `responses/` files only when they materially affect the current decision.
- Do not read `private/` by default; use it only when the user explicitly points to local private context and the files exist locally.
- Never commit private note contents. The `private/` directory should track only `README.md` and `.gitignore`.
- Do not store secrets, tokens, credentials, private keys, or machine-local config in notes, including private notes.
- Promote durable decisions, requirements, constraints, accepted assumptions, and verification evidence into `SPEC.md`, `docs/CONSTITUTION.md`, or durable references.
- Record specific note files that materially shape a feature in front matter `references`.
- Mark stale or superseded note references with `status: stale` and `read_policy: skip` instead of deleting useful history.

## Anti-Patterns

- Treating `docs/notes/<feature>` as a replacement for `SPEC.md`.
- Leaving durable customer requirements only in Slack excerpts, notes, or chat transcripts.
- Loading every file under `docs/notes/` before the immediate decision requires it.
- Committing raw private conversation history or private note contents.
- Using `.gitkeep` files as evidence that a notes directory contains useful context.
- Copying large raw transcripts into canonical docs instead of extracting decision-shaping facts.

## Verification

- Relevant note files are referenced when they materially shaped the work.
- Durable conclusions were promoted to canonical project artifacts.
- Private note contents were not staged.
- `.gitkeep` placeholders were ignored.
- Stale note references were marked stale instead of treated as current inputs.

## Examples

Create a feature notes directory:

```bash
kit notes my-feature
```

Capture a customer ask:

```bash
kit notes my-feature --add --source slack --title "Customer ask"
```

Capture local-only private context:

```bash
kit notes my-feature --add --private --title "Slack conversation"
```

Reference a note that shaped a feature:

```yaml
references:
  - id: customer-ask
    name: Customer ask
    type: notes
    target: docs/notes/0001-my-feature/inbox/2026-06-29-120000-customer-ask.md
    relation: informs
    read_policy: conditional
    used_for: requirement context
    status: active
```
