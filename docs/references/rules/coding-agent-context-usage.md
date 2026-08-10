---
kind: ruleset
slug: coding-agent-context-usage
description: Requires coding agents to resolve and progressively load repository-local workflow, rule, specification, strategy, and implementation evidence.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - context
  - workflow
  - implementation
  - maintenance
  - pull-request
read_policy_default: must
---

# Ruleset: Coding Agent Context Usage

## Purpose

- Make repository-local evidence the normal starting point for coding-agent work.
- Keep context loading deterministic, progressive, and proportional to the immediate decision.
- Separate Kit's evidence resolution from model inference or agent execution.

## Applies When

- A coding agent bootstraps, implements, maintains, reviews, repairs, or coordinates work in a Kit-managed repository.
- Task scope changes enough that the previously resolved workflow or evidence set may no longer be sufficient.

## Rules

1. Use `kit capabilities <command> --json` before choosing a Kit command when its safety or side effects are not already established.
2. Run `kit context resolve --workflow <slug> --json`, adding `--feature` and relevant `--path` hints when applicable.
3. Load the selected required workflow, rules, specifications, strategies, and implementation-pattern evidence before acting.
4. Load conditional evidence only when the current decision reaches its applicability boundary.
5. Re-resolve after a material scope, workflow, feature, path, delivery, or environment change.
6. Treat missing required evidence as blocked implementation context. Repair or explicitly resolve the repository evidence gap before proceeding.
7. Keep Kit local and deterministic: do not ask it to infer project truth, choose a model, launch an agent, or fetch context through `kit context resolve`.

## Anti-Patterns

- Loading every repository document up front.
- Guessing command behavior from memory instead of using capabilities metadata.
- Treating resolved JSON as a new source of truth instead of a projection of repository-local Markdown.
- Continuing after required evidence becomes missing, invalid, or out of scope.
- Copying complete rules or workflow bodies into always-loaded provider entrypoints.

## Verification

- Confirm the resolved contract is `kit.context/v1`, contains only project-relative paths, and is stable for unchanged repository state.
- Confirm every required selected artifact exists, remains inside the repository, and has the reported digest.
- Confirm the agent loaded the required evidence before implementation or mutation.
- Confirm material scope changes caused a new resolution.

## Examples

```bash
kit capabilities context --json
kit context resolve --workflow implementation-delivery --feature invitation-flow --json
kit context resolve --workflow repository-maintenance --json
kit context resolve --workflow pr-feedback-repair --path pkg/service/orders.go --json
```
