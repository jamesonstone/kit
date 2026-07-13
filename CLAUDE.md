# CLAUDE

## Purpose

- This file is a routing table, not the full manual
- Start at `docs/agents/README.md`, then load only the docs needed for the current decision
- Repo-local markdown under `docs/` is the system of record

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

## Runtime Routing

- `docs/agents/README.md` — classify the task and choose the next document
- `docs/agents/WORKFLOWS.md` — spec-driven versus ad hoc flow
- `docs/agents/GUARDRAILS.md` — completion, safety, and hard rules
- `docs/agents/RLM.md` — just-in-time context loading when broad context would be noisy
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary inputs

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before any GitHub delivery mutation, load `docs/agents/GUARDRAILS.md` and the relevant `docs/references/rules/*` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract

## AWS Context Hard Gate

- If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Use the verified configured profile explicitly for every AWS-dependent command, including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts, where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Treat the verified account and ARN as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

## Conditional Context

- `docs/specs/<feature>/` — active feature artifacts only
- `docs/references/README.md` — durable repo references only when relevant
- `docs/CONSTITUTION.md` — project invariants when a decision depends on them

## Repo Knowledge Map

- `docs/agents/README.md` — runtime routing index
- `docs/agents/WORKFLOWS.md` — work classification and source-of-truth semantics
- `docs/agents/RLM.md` — progressive disclosure and context budget rules
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary global inputs
- `docs/agents/GUARDRAILS.md` — completion bar, safety rules, and validation expectations
- `docs/references/README.md` — durable repo-local references that are broader than one feature
- `docs/specs/<feature>/SPEC.md` — v2 feature source of truth for requirements, plan, tasks, validation, reflection, delivery, and evidence

## Constraints

- Keep CLAUDE short and stable so it fits easily into injected context
- Put durable workflow guidance in `docs/agents/*` rather than expanding this file
- Do not add an always-loaded monolithic instruction file
