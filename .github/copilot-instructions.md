# GitHub Copilot Repository Instructions

## Quick Rules

- Use this file as a map, not the full manual
- Start with `docs/agents/README.md` and then open only the linked docs needed for the current decision
- Treat `docs/specs/<feature>/SPEC.md` as the v2 feature system of record
- Use `docs/agents/RLM.md` when full-context loading would be noisy or wasteful
- Keep context minimal and source-attributed

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

## Runtime Routing

- `docs/agents/README.md` — classify the task and choose the next document
- `docs/agents/WORKFLOWS.md` — workflow and source-of-truth rules
- `docs/agents/GUARDRAILS.md` — hard completion and safety rules
- `docs/agents/RLM.md` — just-in-time context routing
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary inputs

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before any GitHub delivery mutation, load `docs/agents/GUARDRAILS.md` and the relevant `docs/references/rules/*` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract

## AWS Context Hard Gate

- If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Treat the verified account and ARN as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

## Non-Negotiable Rules

- Repo-local docs under `docs/` are the source of truth
- Always update affected documentation and keep touched docs properly formatted
- Keep context minimal and load only the docs and files relevant to the task
- Remove dead code and unnecessary exports or public surface when they are not strictly needed
- Do not treat `.claude/skills` as canonical discovery input
- Do not add an always-loaded monolithic instruction file

## Repo Knowledge Map

- `docs/agents/README.md` — repo-local entrypoint
- `docs/agents/WORKFLOWS.md` — work classification and execution flow
- `docs/agents/RLM.md` — progressive-disclosure pattern for broad discovery
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary globals
- `docs/agents/GUARDRAILS.md` — hard rules and completion bar
- `docs/references/README.md` — durable repo-local references
- `docs/specs/<feature>/SPEC.md` — v2 feature source of truth
