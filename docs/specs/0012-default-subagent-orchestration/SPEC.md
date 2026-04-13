# SPEC

## SUMMARY

- Change Kit's shared prompt-orchestration default from single-agent to subagent-first.
- Add `--single-agent` as the explicit opt-out when a user wants one lane only.
- Keep `kit dispatch` as the stricter discovery-first queue-planning command with explicit approval before launch.
- Clarify that repository-scale RLM discovery narrows context first, while dispatch and subagents handle execution planning after discovery.

## PROBLEM

- Kit currently assumes a single-agent lane unless the user explicitly adds `--subagents`.
- That default underuses subagents on work that naturally splits across multiple distinct areas in both research and implementation.
- Users now want subagent orchestration to be the standard path, while still preserving conservative overlap handling and a way to force one-lane execution when needed.

## GOALS

- Make shared prompt augmentation include subagent orchestration guidance by default.
- Add `--single-agent` as a root-level opt-out for all prompt-producing commands that use the shared prompt helper.
- Preserve the current safety rules around overlap prediction, conservative clustering, and main-agent integration.
- Keep `kit dispatch` as the formal queue-planning command for high-ambiguity or approval-gated orchestration.
- Update README, help, and tests so the shipped behavior is explicit.

## NON-GOALS

- Launching subagents directly from the Kit binary.
- Removing `kit dispatch`.
- Relaxing `kit dispatch` approval gating.
- Changing command-specific planning or execution rules outside the shared orchestration suffix.

## USERS

- Users who want prompt outputs to default to parallel, subagent-capable workflows.
- Users who still occasionally need a forced one-lane prompt.
- Coding agents that need a consistent default orchestration model across research and implementation prompts.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| shared prompt helper | code | /Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/subagents.go | control the default orchestration suffix | active |
| dispatch command spec | doc | /Users/jamesonstone/go/src/github.com/jamesonstone/kit/docs/specs/0008-dispatch-command/SPEC.md | preserve the stricter queue-planning role of `kit dispatch` | active |

## REQUIREMENTS

- [SPEC-01] The shared prompt helper must append `## Subagent Orchestration` by default for prompt-producing commands that use `prepareAgentPrompt(...)`.
- [SPEC-02] A new root persistent flag `--single-agent` must disable the shared subagent-orchestration suffix.
- [SPEC-03] The shared orchestration suffix must instruct the coding agent to default to subagents when the work spans multiple distinct areas.
- [SPEC-04] The shared orchestration suffix must still require predicting likely touched files or interfaces before parallelizing work.
- [SPEC-05] The shared orchestration suffix must still require conservative overlap clustering and avoidance of unsafe parallelism.
- [SPEC-06] The shared orchestration suffix must instruct the coding agent to fall back to `kit dispatch`-style discovery discipline when overlap or ambiguity is high.
- [SPEC-07] The shared orchestration suffix must keep the main agent responsible for synthesis, integration, validation, and communication.
- [SPEC-08] `kit dispatch` must keep using the no-shared-subagent-suffix output path so its dedicated prompt remains the only orchestration guidance in that command.
- [SPEC-09] README and help output must describe the new default as “subagents by default” and `--single-agent` as the explicit opt-out.
- [SPEC-10] The legacy `--subagents` flag may remain only as a hidden compatibility alias and must not be the documented primary interface.
- [SPEC-11] Shared subagent guidance must distinguish repository-scale discovery from execution planning:
  - RLM narrows candidate docs, files, and workstreams first
  - dispatch or subagent execution begins only after that scope is narrowed enough to predict overlap conservatively

## ACCEPTANCE

- A normal prompt-producing command output includes `## Subagent Orchestration` without requiring any extra flag.
- Running the same command with `--single-agent` omits that shared orchestration section.
- The shared subagent section tells the coding agent to default to subagents across distinct areas while still predicting touched files, clustering overlap conservatively, and keeping the main agent responsible for integration.
- The shared subagent section explains that RLM is discovery-first context narrowing, while dispatch and subagent execution happen only after overlap can be predicted conservatively.
- `kit dispatch` output remains free of the shared subagent suffix and continues to use its stricter discovery-first queue-planning prompt.
- README and tests reflect the new default and the `--single-agent` escape hatch.

## EDGE-CASES

- A user explicitly wants a one-lane prompt for narrow or risky work.
- A user still passes the old `--subagents` flag from existing scripts.
- Dispatch output must remain unchanged even though subagents are now the shared default elsewhere.
- The work is broad enough for subagents, but overlap confidence is low and must stay conservative.

## OPEN-QUESTIONS

- none
