package cli

func v2GuidanceExpectations() map[string][]string {
	return map[string][]string{
		"docs/agents/README.md": {
			"## Runtime Routing",
			"load only the linked doc needed for the current decision",
			"Stop loading once the decision is supported",
		},
		"docs/agents/RLM.md": {
			"## Runtime Loop",
			"identify the immediate decision",
			"load the smallest relevant artifact",
			"stop loading once the decision is supported",
			"## Context Budget Rules",
			"specific section over full file",
			"repo-local docs before global model/vendor instructions",
			"`docs/references/rules/testing-and-environment-validation.md`",
			"`docs/references/testing.md`",
		},
		"docs/agents/WORKFLOWS.md": {
			"Authority order:",
			"Execution order for feature work:",
			"`SPEC.md` controls requirements, plan, tasks, validation, reflection, delivery, and evidence",
			"`BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` are non-binding historical context in v2",
		},
		"docs/agents/GUARDRAILS.md": {
			"Never claim tests passed unless they ran",
			"Never claim files were inspected unless they were inspected",
			"If validation cannot run, state why",
			"docs/references/rules/source-file-size.md",
			"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		},
		"docs/references/README.md": {
			"`rules/testing-and-environment-validation.md`",
			"`rules/source-file-size.md`",
		},
		"docs/references/testing.md": {
			"`rules/testing-and-environment-validation.md`",
			"## Code-Level Validation",
			"## High-Level Suites",
			"## Environment Preflights",
			"## Evidence And Retention",
			"## Known Gaps",
		},
	}
}

func v3GuidanceExpectations() map[string][]string {
	return map[string][]string{
		".github/copilot-instructions.md": {
			"Before implementation or validation, load `docs/references/rules/testing-and-environment-validation.md`",
			"end-to-end and live-integration suites supplement rather than replace them",
			"Before editing implementation/source or test files, load `docs/references/rules/source-file-size.md`",
			"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		},
		"docs/agents/README.md": {
			"## Runtime Routing",
			"Native agent planning owns research, clarification, design, and plan formation",
			"V1 and V2 artifacts remain supported legacy inputs",
		},
		"docs/agents/WORKFLOWS.md": {
			"## Native Planning To Repository Memory",
			"Before code, assess whether the work contains material rationale",
			"Never mechanically rewrite a V2 spec into V3",
			"`kit dispatch` supports post-plan execution topology",
		},
		"docs/agents/RLM.md": {
			"## Runtime Loop",
			"identify the immediate decision",
			"stop loading once the decision is supported",
			"## Context Budget Rules",
			"`docs/references/rules/testing-and-environment-validation.md`",
			"`docs/references/testing.md`",
		},
		"docs/agents/TOOLING.md": {
			"Use `kit dispatch` after native planning",
			"accepted plan needs a safe multi-lane execution topology",
			"Link the primary checkout's `.env` and `.envrc` into writable lanes by default",
			"preserve a repository- or user-supplied `.envrc`",
		},
		"docs/agents/GUARDRAILS.md": {
			"## Repository Memory Completion Gate",
			"Create or adopt a spec before code when material rationale exists",
			"Every implementation final response must include `Repository Memory`",
			"docs/references/rules/source-file-size.md",
			"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		},
		"docs/references/README.md": {
			"`rules/testing-and-environment-validation.md`",
			"`rules/source-file-size.md`",
			"`worktrees.md` for the canonical native Git worktree hierarchy",
			"environment ownership",
		},
		"docs/references/testing.md": {
			"`rules/testing-and-environment-validation.md`",
			"## Code-Level Validation",
			"## High-Level Suites",
			"## Environment Preflights",
			"## Evidence And Retention",
			"## Known Gaps",
		},
		"docs/references/worktrees.md": {
			"optional `git wt` helper opens the same colorized selector",
			"The `PR#` column runs one batched `gh` lookup",
			"For direct branch navigation, use `git wt <branch>`",
			"## Writable-Lane Environment Links",
			"Link each stable source into writable lanes by default",
			"Preserve a regular destination",
			"`.envrc`, which may be tracked by Git or owned by the user",
			"Verified `.env` and `.envrc`",
		},
	}
}
