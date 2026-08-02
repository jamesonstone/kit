---
kind: ruleset
slug: codex-thread-initialization
description: Requires ordered pre-response Codex task renaming and pinning with verified or fail-visible status.
status: active
registry_scope: downstream
applies_to:
  - codex
  - coding-agent
  - session
  - thread
  - session-management
read_policy_default: must
---

# Ruleset: Codex Thread Initialization

## Purpose

- Make task identity and visibility deterministic at the start of every newly
  created Codex task.
- Replace optional best-effort wording with an ordered, observable pre-response
  contract.
- Preserve progress when a host capability is genuinely unavailable without
  allowing silent omission.

## Applies When

- Codex begins a newly created task in a repository that carries this rule.
- Kit generates, refreshes, or reconciles V3 `AGENTS.md` instructions.
- A continued task is missing its expected title or pin state, or the user asks
  to change either state.

The root `AGENTS.md` gate is self-contained because initialization must happen
before repository inspection. Load this detailed rule when generating,
refreshing, reconciling, or reviewing the instruction contract.

## Rules

### Pre-Response Boundary

- Treat initialization as a blocking pre-response gate for a newly created
  Codex task.
- Complete the gate before the first commentary message, planning, repository
  inspection, shell or network commands, or any other substantive task action.
- Before the gate, permit only the minimum capability lookup required to locate
  the host's thread-title and thread-pin operations.

### Required Order

1. Invoke the available thread-title operation, such as `set_thread_title`,
   with `[<project>] <description>`.
2. Invoke the available thread-pin operation, such as `set_thread_pinned`.
3. Verify each result from returned host state when that state is available.

Both operations are required and ordered. Do not defer a supported operation
to a later interaction.

### Title Contract

- Derive `<project>` from repository or working-directory context supplied by
  the host. Do not inspect the repository merely to choose the title.
- Derive `<description>` from the user's request.
- Keep the description lowercase and at most four words.
- Do not let title wording delay the required operations; choose the clearest
  concise description supported by the supplied context.

### Failure And Continuation

- Distinguish success, unsupported or unavailable capability, and operation
  failure. Do not collapse them into an unverified “attempted” state.
- Resolve both actions in order even when the first one cannot succeed.
- Do not silently skip an operation or retry it indefinitely.
- When either result is not successful, begin the first commentary with
  `Thread initialization: rename <status>; pin <status>.` and include a concise
  reason for each non-success status.
- After the fail-visible status, continue the user's substantive request. A
  missing host capability does not authorize abandoning unrelated work.

### Continued Tasks

- Preserve the title and pin state of a continued Codex task.
- Re-run only the missing operation unless the user explicitly requests a title
  or pin-state change.
- Do not repeat successful initialization on every interaction.

### Provider Boundary

- Keep this behavior in Codex-facing `AGENTS.md` instructions.
- Do not copy Codex task-state operations into Claude or GitHub Copilot
  instructions.

## Anti-Patterns

- Saying only “attempt to rename and pin” without required ordering or status.
- Emitting a greeting, plan, or inspection update before initialization.
- Inspecting the repository before choosing a title when host context already
  provides the project name.
- Deferring rename or pin until after substantive work begins.
- Silently treating an unavailable operation as success.
- Retrying a missing capability until the user's actual task is blocked.
- Renaming or pinning an already initialized continued task on every message.

## Verification

- Confirm the hard gate is the first section after the `AGENTS.md` title.
- Confirm title invocation precedes pin invocation.
- Confirm no substantive task action is authorized before both results resolve.
- Confirm returned state is checked when available.
- Confirm a non-success result has the exact fail-visible first-commentary
  format and does not block subsequent substantive work.
- Confirm generated Claude and Copilot instructions do not contain the gate.
- Confirm `kit reconcile` reports missing or weakened gate semantics in an
  existing V3 `AGENTS.md`.

## Examples

Successful initialization:

```text
set_thread_title("[payments] retry policy") -> renamed
set_thread_pinned(true) -> pinned
first commentary follows
```

Unavailable title operation:

```text
thread-title capability lookup -> unavailable
set_thread_pinned(true) -> pinned
Thread initialization: rename unavailable: host operation absent; pin pinned.
```
