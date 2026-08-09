---
kind: ruleset
slug: feature-notes
description: Treats docs/notes feature material as optional input while keeping durable truth in canonical repository documentation.
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

# Ruleset: Feature Notes

## Purpose

Use raw research, conversation excerpts, screenshots, and draft material as
optional context without treating it as authoritative repository truth.

## Applies When

- The task or a spec reference names `docs/notes/<feature>` or another
  repository-local source-material path.
- A decision depends on raw context not yet represented in canonical docs.

## Rules

- List the smallest relevant note set before reading file contents.
- Ignore empty placeholders and do not load every note by default.
- Treat notes as evidence, not requirements.
- Promote durable requirements, decisions, constraints, and validation into
  the relevant `SPEC.md`, `docs/CONSTITUTION.md`, domain documentation, or
  reusable reference.
- Record materially used note paths in spec references.
- Do not read local-private material unless the user explicitly identifies it.
- Never commit secrets, credentials, private keys, or private conversation
  content.
- Mark superseded references stale rather than silently treating them as
  current.

## Anti-Patterns

- Replacing a specification with raw notes.
- Leaving a durable customer requirement only in a transcript.
- Copying large source transcripts into canonical documentation.
- Treating an empty notes directory as relevant context.

## Verification

- Every materially used source path is identifiable.
- Durable conclusions were promoted to the correct canonical artifact.
- No private or secret material was staged.

## Examples

```yaml
references:
  - id: customer-ask
    name: Customer ask
    type: notes
    target: docs/notes/0058-coding-agent-first/customer-ask.md
    relation: informs
    read_policy: conditional
    used_for: requirement context
    status: active
```
