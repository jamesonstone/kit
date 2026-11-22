---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0105"
  slug: "harness-research-notes"
  dir: "0105-harness-research-notes"
references:
  - id: harness-research-notes
    name: Kit research publishing rule
    type: ruleset
    target: docs/references/rules/project/harness-research-notes.md
    relation: constrains
    read_policy: must
    used_for: Delivering Kit changes with evidence and reasoning
    status: active
---
# SPEC

## PURPOSE

Make the reasoning and evidence behind changes to Kit available as short research posts.

## CONTEXT

The first post covers Kit #304 and #306. The project owner wants future Kit changes to carry the same record without giving downstream projects a publishing obligation.

## REQUIREMENTS

- Cover features, fixes, refactors, tweaks, and design changes under the series title *Coding-Agent Harness Research*.
- Deliver the post through the website's existing review and publication workflow. Distinguish a prepared post from a live one.
- Make the rule discoverable in Kit's own agent instructions. Keep it outside the binary and downstream scaffolds.

## ACCEPTED PLAN

Add a project-owned ruleset below the embedded rule directory and link it from the project guidance in the three agent entry files. Pair the rule with the first site post in [site issue #118](https://github.com/jamesonstone/jamesonstone.github.io/issues/118).

## DECISIONS

A single post can cover a coherent set of changes when every change in the set is described. Linking the source and website PRs preserves a publication obligation across the two repositories without adding a release gate, deployment job, or CLI feature.

## DISCOVERIES

The Go embed pattern matches top-level rule Markdown. A project subdirectory keeps the source rule physically outside the binary while retaining its ruleset format and an explicit local trigger.

## VALIDATION

- Project and feature validation passed. The project check retained the existing references compatibility advisory.
- Managed instruction parity and the rest of `go test ./internal/templates` passed. `go list` confirmed the rule is outside the Go embed set. A downstream init preview included the shared delivery rule and omitted the project rule.
- The site build rendered the companion permalink and passed the site validator. The site pull request records publication separately.

## OUTCOME

The Kit rule defines when to draft, review, publish, and verify a post. The website source and public status are tracked separately through the linked site issue.

## REPOSITORY MEMORY

The project-only ruleset owns publishing procedure. The Kit Constitution records the scope boundary. Published posts preserve public reasoning and evidence.
