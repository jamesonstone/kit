---
kind: ruleset
slug: agent-team-orchestration
description: Keeps delegated agent work safe, integrated, and truthfully reported without prescribing a planning lifecycle.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - workflow
  - dispatch
  - subagent
  - verification
read_policy_default: conditional
---

# Ruleset: agent-team-orchestration

## Purpose

- Let agents use their host's native subagents and parallelism wherever that improves correctness or speed.
- Keep one accountable primary agent and prevent conflicting writes.
- Keep reports truthful about what actually ran.

Kit does not inspect agent rosters, choose models, launch agents, or supervise them. The active coding agent decides topology with the controls its host actually exposes.

## Applies When

- A coding agent delegates work to subagents or parallel workers, including through `kit dispatch` or `kit pr fix` prompts.
- A report claims parallel execution or independent verification.

## Rules

- Decompose when independent investigation, implementation, or verification lanes clearly help; keep tightly coupled or small work in one agent. No recorded topology decision or manifest is required.
- The primary agent owns scope, integration, validation, delivery, and the final report. Delegated agents never create issues, branches, commits, pushes, pull requests, comments, or merges.
- Never let two agents write overlapping files concurrently; serialize shared files and integrate one coherent diff.
- Give each delegated agent a self-contained brief: goal, scope, files it may change, constraints, and the evidence to return.
- Verify delegated results against the current source before integrating them.
- Report only real topology: never describe a logical lane, a task list, or supervisor self-review as a separate agent or independent verification. When a requested control (a specific model, parallelism, or fresh verifier) is unavailable, say so and continue in the best available configuration.

## Anti-Patterns

- Writing manifests, lifecycle states, or topology records that no one consumes.
- Parallel agents editing the same file.
- A delegated agent committing, pushing, or commenting.
- Claiming independent verification that the primary agent performed itself.

## Verification

- Every integrated change came through the primary agent's diff and validation.
- The final report states which work ran in separate agents only when it did.

## Examples

```text
Three independent read-only investigations run as subagents; the primary agent
integrates their findings, implements the change, and runs validation.
```

```text
A one-file fix stays in the primary agent; no delegation or topology record.
```
