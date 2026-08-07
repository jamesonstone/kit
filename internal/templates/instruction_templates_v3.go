package templates

import "strings"

func memoryRepositoryInstructions(title string) string {
	return codexThreadInitializationGate(title) + codexBrowserPolicy(title) + `## Purpose

- This file is a routing table, not the full manual
- Start at ` + "`docs/agents/README.md`" + ` and load only the guidance needed for the current decision
- Use native agent planning for research, clarification, design, and implementation planning
- Treat repo-local markdown under ` + "`docs/`" + ` as persistent repository memory

## Repository Memory Gate

- Before implementation, inspect relevant code and existing repository memory
- Decide semantically whether the work contains material rationale that code and tests cannot preserve
- When material rationale exists, create or adopt ` + "`docs/specs/<feature>/SPEC.md`" + ` before editing implementation files and capture the accepted native plan
- When code and tests are sufficient, do not create documentation solely to satisfy a process; record ` + "`not required`" + ` in the final Repository Memory report
- During implementation, keep material decisions and discoveries current in the spec
- After implementation and validation, load ` + "`docs/references/rules/constitution-curation.md`" + `; curate feature rationale into ` + "`SPEC.md`" + `, demonstrated project invariants into ` + "`docs/CONSTITUTION.md`" + `, reusable practices into ` + "`docs/references/`" + ` or ` + "`docs/references/rules/`" + `, and domain knowledge into its existing canonical documentation
- Remove transient planning chatter and code-recoverable detail during curation; retain material superseded decisions with rationale

## Final Response Contract

- Every implementation final response must include:
  - ` + "`Repository Memory`" + `
  - ` + "`Decision: created | updated | refactored | deleted | not required`" + `
  - ` + "`Rationale: <why this is the correct persistence decision>`" + `
  - ` + "`Artifacts: <paths or none>`" + `

## Runtime Routing

- ` + "`docs/agents/README.md`" + ` — classify the work and choose the next document
- ` + "`docs/agents/WORKFLOWS.md`" + ` — native planning, implementation, and repository-memory lifecycle
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion, safety, and hard rules
- ` + "`docs/agents/RLM.md`" + ` — just-in-time context loading
- ` + "`docs/agents/TOOLING.md`" + ` — skills, post-plan dispatch, and secondary inputs

## Testing And Validation Gate

- Before implementation or validation, including browser automation and browser testing, load ` + "`docs/references/rules/testing-and-environment-validation.md`" + ` and the project's ` + "`docs/references/testing.md`" + `
- Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them

## Source File Size Gate

- Before editing implementation/source or test files, load ` + "`docs/references/rules/source-file-size.md`" + `
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less
- Audit the complete affected source/test scope before delivery; whole-project reconcile and scheduled maintenance audit the entire repository

## Application Architecture Gate

- Before implementing API or backend routes, controllers or handlers, services, repositories, persistence adapters, or gateways, load ` + "`docs/references/rules/backend-service-architecture.md`" + `
- Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load ` + "`docs/references/rules/frontend-application-architecture.md`" + `
- Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture

## GitHub Delivery Hard Gate

- Issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before a delivery mutation, load ` + "`docs/agents/GUARDRAILS.md`" + ` and relevant ` + "`docs/references/rules/*`" + ` delivery rules
- Repo-local Kit rules outrank generic GitHub or plugin defaults

## AWS Context Hard Gate

- If ` + "`.kit.yaml`" + ` defines an enabled AWS context, run ` + "`kit aws verify`" + ` before the first AWS-dependent command and again immediately before AWS mutation
- Use only the verified configured profile; stop on missing credentials, incomplete configuration, or identity mismatch

## Knowledge Map

- ` + "`docs/specs/<feature>/SPEC.md`" + ` — material feature rationale and living implementation history
- ` + "`docs/CONSTITUTION.md`" + ` — project invariants
- ` + "`docs/references/`" + ` — reusable repo-wide knowledge and practices
- domain documentation — canonical domain behavior and interfaces
- ` + "`docs/notes/<feature>/`" + ` — optional source material, never canonical truth by itself

## Constraints

- Keep ` + title + ` short and stable
- Put durable workflow guidance in ` + "`docs/agents/*`" + ` instead of expanding always-loaded files
- Do not ingest or depend on agent transcripts as repository memory
`
}

func codexThreadInitializationGate(title string) string {
	if title != "AGENTS" {
		return ""
	}

	return `## Codex Thread Initialization Hard Gate

- For every newly created Codex task, this is a blocking pre-response gate. Complete it before the first commentary message and before planning, repository inspection, shell or network commands, or any other task action; only the minimum capability lookup needed to locate the two thread operations may precede it.
- First, call the available thread-title operation (` + "`set_thread_title`" + ` when available) with ` + "`[<project>] <description>`" + `.
- Second, call the available thread-pin operation (` + "`set_thread_pinned`" + ` when available).
- Both actions are required and ordered. Never defer either supported operation to a later interaction.
- Derive ` + "`<project>`" + ` from the host-provided repository or working-directory context and ` + "`<description>`" + ` from the user request without inspecting the repository first. Keep the description lowercase and at most four words.
- Verify each operation from its returned state when the host exposes one.
- If an operation is unsupported, unavailable, or fails, do not silently skip it or retry indefinitely. After resolving both actions in order, begin the first commentary with ` + "`Thread initialization: rename <status>; pin <status>.`" + `, include a concise reason for every non-success status, then continue the requested work.
- For a continued Codex task, preserve its current title and pin state unless either is missing or the user explicitly requests a change.

`
}

const memoryCopilotInstructions = `# GitHub Copilot Repository Instructions

## Native Planning

` + `Use native planning for research and design. Before implementation, inspect code and repository documentation, then decide whether material rationale requires a living ` + "`SPEC.md`" + `. Capture the accepted plan before code when it does. After validation, load ` + "`docs/references/rules/constitution-curation.md`" + ` and curate durable decisions into the correct repository document; code-and-test-sufficient work may report that no documentation update was required.

Start with ` + "`docs/agents/README.md`" + `. Before implementing API or backend routes, handlers, services, repositories, persistence adapters, or gateways, load ` + "`docs/references/rules/backend-service-architecture.md`" + `. Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load ` + "`docs/references/rules/frontend-application-architecture.md`" + `. Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture.

Before implementation or validation, including browser automation and browser testing, load ` + "`docs/references/rules/testing-and-environment-validation.md`" + ` and the project's ` + "`docs/references/testing.md`" + `. Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them.

Before editing implementation/source or test files, load ` + "`docs/references/rules/source-file-size.md`" + `. Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less, audit the complete affected scope before delivery, and audit the entire repository during whole-project reconcile and scheduled maintenance.

Before Git, GitHub, or AWS mutations, load ` + "`docs/agents/GUARDRAILS.md`" + ` and relevant ` + "`docs/references/rules/*`" + `. Repo-local Kit rules outrank generic defaults.

## Final Response

Every implementation final response must include:

- Repository Memory
- Decision: created | updated | refactored | deleted | not required
- Rationale: why this persistence decision is correct
- Artifacts: paths or none
`

func memoryInstructionSupportContent(relativePath string) string {
	switch relativePath {
	case "docs/agents/README.md":
		return `# Agents Docs

## Purpose

- Route agents from native planning through implementation to curated repository memory
- Load only the guidance and repository context needed for the current decision

## Runtime Routing

- ` + "`WORKFLOWS.md`" + ` — native-plan lifecycle and memory routing
- ` + "`GUARDRAILS.md`" + ` — safety, completion, validation, and final-response rules
- ` + "`RLM.md`" + ` — progressive disclosure for broad context
- ` + "`TOOLING.md`" + ` — skills, execution topology, and secondary inputs
- ` + "`docs/specs/<feature>/SPEC.md`" + ` — material feature rationale when required
- ` + "`docs/references/`" + ` — durable reusable knowledge

## System Of Record

- Native agent planning owns research, clarification, design, and plan formation
- The repository owns durable rationale; chat and transcripts do not
- V3 ` + "`SPEC.md`" + ` records purpose, context, requirements, accepted plan, decisions, discoveries, validation, outcome, and repository-memory curation
- V1 and V2 artifacts remain supported legacy inputs and must not be mechanically rewritten into V3
`
	case "docs/agents/WORKFLOWS.md":
		return `# Workflows

## Native Planning To Repository Memory

1. Inspect the request, relevant code, and existing repository memory.
2. Use the host agent's native planning capability for research, clarification, design, and implementation planning.
3. Before code, assess whether the work contains material rationale that code and tests cannot preserve.
4. When it does, create or adopt ` + "`docs/specs/<feature>/SPEC.md`" + ` and translate the accepted native plan into it before implementation.
5. Keep material decisions and discoveries current while implementing.
6. Validate the implementation.
7. Load ` + "`docs/references/rules/constitution-curation.md`" + ` and curate the spec and broader repository memory to match what was actually built.

` + "`kit spec [feature]`" + ` scaffolds or adopts the living spec and provides orientation. It does not replace native planning and does not ingest transcripts. The legacy V2 supervisor is compatibility-only.

## Memory Decision

- Create or update a spec for consequential product behavior, architecture, cross-component contracts, rejected alternatives, or historical decisions future agents need.
- Do not create a spec for mechanical or code-sufficient work when code and tests communicate the complete durable truth.
- Route feature rationale to ` + "`SPEC.md`" + `, invariants to ` + "`CONSTITUTION.md`" + `, reusable practices to references or rules, and domain knowledge to existing canonical domain docs.
- Treat the exact generated Constitution starter as a valid bootstrap state; promote only demonstrated project-wide truth through the Constitution curation rule.

## V3 Phase Gates

- Before implementation: purpose, context, requirements including non-goals and observable acceptance, and accepted plan must be populated.
- At completion: decisions and discoveries must be resolved, validation and actual outcome recorded, repository memory assessed, and pending placeholders removed.

## Compatibility

- V1 and V2 specs remain readable and valid.
- Never mechanically rewrite a V2 spec into V3; migration requires semantic curation.
- Bare ` + "`kit loop`" + ` and ` + "`kit loop workflow`" + ` are deprecated V2 compatibility paths. V3 work uses native planning.
- ` + "`kit dispatch`" + ` supports post-plan execution topology; it does not design the feature.
`
	case "docs/agents/GUARDRAILS.md":
		return memoryGuardrails()
	case "docs/agents/RLM.md":
		return strings.ReplaceAll(strings.ReplaceAll(agentsRLM, "For v2 feature-scoped work", "For living-spec feature work"), "Use `kit dispatch` only when the work moves from broad discovery into multi-lane execution planning", "Use `kit dispatch` only after native planning has established a narrow implementation topology")
	case "docs/agents/TOOLING.md":
		return memoryTooling()
	case "docs/references/testing.md":
		return strings.ReplaceAll(referencesTesting, "Validation Map and Evidence sections", "VALIDATION and OUTCOME sections")
	case "docs/references/README.md":
		return strings.ReplaceAll(
			referencesREADME,
			"Use `worktrees.md` when present for the canonical native Git worktree hierarchy, naming, shared-state model, safety contract, and optional manual convenience commands",
			"Use `worktrees.md` for the canonical native Git worktree hierarchy, naming, shared-state model, environment ownership, and safety contract",
		)
	case "docs/references/worktrees.md":
		return referencesWorktrees
	default:
		return ""
	}
}

func memoryGuardrails() string {
	content := strings.ReplaceAll(agentsGuardrails, "For v2 feature work, populate all required `SPEC.md` sections and keep front matter `workflow_version`, `phase`, references, relationships, and skills current", "For V3 feature work, satisfy the phase-aware living-spec gates and keep front matter `workflow_version`, `phase`, references, relationships, and skills current; preserve version-specific requirements for legacy specs")
	content = strings.ReplaceAll(content, "Link the invoking checkout's `.env`", "Link the primary checkout's `.env`")
	return content + `

## Repository Memory Completion Gate

- Inspect existing repository memory before implementation.
- Create or adopt a spec before code when material rationale exists.
- After implementation and validation, curate durable rationale into the correct canonical documents.
- A justified ` + "`not required`" + ` decision is valid when code and tests preserve the complete durable truth.
- Every implementation final response must include ` + "`Repository Memory`" + `, a valid decision, rationale, and artifact paths or ` + "`none`" + `.
`
}

func memoryTooling() string {
	content := strings.ReplaceAll(
		agentsTooling,
		"Use `kit dispatch` when broad work must be turned into a safe Agent Team Plan",
		"Use `kit dispatch` after native planning when an accepted plan needs a safe multi-lane execution topology",
	)
	content = strings.ReplaceAll(
		content,
		"Link the invoking checkout's `.env` into writable lanes by default when it exists, using only an exact verified symlink; omit the link when isolation is required",
		"Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists, using only exact verified symlinks; omit both links when isolation is required",
	)
	content = strings.ReplaceAll(
		content,
		"Never copy `.env` contents or automatically share `.envrc`; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
		"Never copy environment contents or overwrite destination environment material; preserve a repository- or user-supplied `.envrc`, and remember that direnv approval remains path-specific; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
	)
	content = strings.ReplaceAll(
		content,
		"except for preparing the writable worktree and its exact `.env` link when needed",
		"except for preparing the writable worktree and its exact `.env` and `.envrc` links when needed",
	)
	return strings.ReplaceAll(
		content,
		"Load `docs/references/worktrees.md` when present and worktree creation",
		"Load `docs/references/worktrees.md` when worktree creation",
	)
}
