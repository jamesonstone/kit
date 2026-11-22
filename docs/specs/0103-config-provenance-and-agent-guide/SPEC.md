---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0103"
  slug: "config-provenance-and-agent-guide"
  dir: "0103-config-provenance-and-agent-guide"
---
# SPEC

## PURPOSE

Use one configuration schema at both scopes, keep personal defaults out of repository writes, and give coding agents a concise source for choosing Kit commands.

## CONTEXT

GH-303 follows the harness audit. The work-ledger contract and delivery rule published flat keys that the typed configuration ignored. Reconcile planned to serialize the global and project overlay into the repository. The interactive rule authoring prompt and frontend rule referenced retired guidance. The full guide repeated all flags and examples on every text lookup.

## REQUIREMENTS

- Global `$XDG_CONFIG_HOME/kit/.kit.yaml` and repository `.kit.yaml` use the same schema. Repository keys override global defaults. Repository writes preserve ownership.
- A flat work-ledger example produces a repair diagnostic. Config reads do not silently enable a previously inactive provider.
- Rule authoring and frontend guidance point to current sources.
- `kit guide` is the official coding-agent guide. Provide a concise index, focused detail, and deterministic JSON.
- Shipping policy, worktree cleanup, handoff, review gates, and spec format remain outside this change.

## ACCEPTED PLAN

Separate effective config reads from repository config writes. Correct the published example and cover nested overrides, provenance, and legacy keys in tests. Remove the two retired guidance routes. Derive the short and focused guide from the live command tree. Run project validation and request review for GH-303.

## DECISIONS

Use the existing typed ledger schema at both scopes. Reject obsolete flat keys with guidance for correcting them. Use the Unix XDG configuration path for global settings. An absolute nonempty `XDG_CONFIG_HOME` selects the base directory. Otherwise use `~/.config`. Read compatibility defaults from `~/.kit/.kit.yaml`, then conventional `~/.config/kit/.kit.yaml` when distinct from the canonical path. Write only to the canonical file. Keep issue settings repository-local. Keep command detail retrievable by an exact command path and preserve schema v1 for the complete JSON catalog.

## DISCOVERIES

The initial implementation selected `~/.kit/.kit.yaml`. The owner superseded that choice with the Unix XDG convention in PR #304. The CodeRabbit review exposed differing global-provider behavior between config loading and scaffold planning. Both now use repository-owned issue settings, as required by delivery. Follow-up review found that global GitHub repository identities could reach generated README links. README generation now uses a repository setting or its origin, while auto-assignment still accepts user defaults. Reconcile and rule registry writes use repository-owned values. Effective paths drive scaffold planning. Explicit blank and null overrides, and omitted keys, retain their meaning when Kit rewrites YAML. Schema and scaffold versions describe the repository file, so global defaults cannot mark a project as migrated.

## VALIDATION

- Focused fixtures verified global and project precedence, obsolete ledger diagnostics, shipped examples, preview and apply isolation, and explicit blank or null overrides.
- Full Go tests, vet, lint, build, gofmt, diff check, project validation, and this feature check passed locally. Project validation printed the existing references compatibility advisory.
- The built text guide returned 50 lines and 534 words. Focused JSON returned one command. The full schema v1 catalog returned 35 command entries and eight artifact entries.
- Fresh-context review reproduced lost overrides, inherited path drift, and force-reset parity issues. Follow-up fixtures passed on the corrected source.
- XDG follow-up fixtures cover path selection, layered fallbacks, usage opt-out and write isolation, YAML-merged settings, repository-local issue routing, and generated README identity.
- No live ledger API writes were used for config validation.

## OUTCOME

Kit uses the canonical global path and preserves repository ownership through configuration rewrites. Custom rules and frontend guidance point to current sources. The guide defaults to a short catalog, with text or structured command detail on request. Existing shipping policy, worktree cleanup, handoff, review gates, and the spec format are outside the change.

## REPOSITORY MEMORY

This spec records the change. The README, canonical contract, delivery guidance, and command reference will carry the current user-facing behavior.
