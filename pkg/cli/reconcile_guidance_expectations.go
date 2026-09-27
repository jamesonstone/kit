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
			"Load `docs/references/rules/testing-and-environment-validation.md` and `docs/references/testing.md` before implementation or validation, including browser automation and browser testing",
			"Load `docs/references/rules/deadline-mode.md` only when the user explicitly signals a real time constraint or deadline in-thread; never infer or proactively suggest deadline mode",
			"Load `docs/references/rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state",
			"Load `docs/references/rules/slack-read-only.md` before any Slack write, and when a Slack link, channel, thread, or search is part of the task",
			"Load `docs/references/rules/aws-agent-toolkit-guidance.md` before AWS-dependent work",
			"Load `docs/references/rules/infrastructure-change-approval.md` before planning or performing mutations to public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state",
			"Load `docs/references/rules/cross-repository-program-coordination.md` before implementing or resuming an accepted plan that spans multiple repositories with dependent deliverables, staged deployment or activation, or expected agent or session handoff",
		},
		"docs/agents/TOOLING.md": {
			"When `cross-repository-program-coordination` applies, dispatch only the canonical program ledger's reconciled ready frontier and checkpoint program state after each material transition or handoff",
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
		},
		"docs/references/README.md": {
			"`rules/deadline-mode.md`",
			"`rules/aws-agent-toolkit-guidance.md`",
			"`rules/deletion-safety.md`",
			"`rules/slack-read-only.md`",
			"`rules/infrastructure-change-approval.md`",
			"`rules/cross-repository-program-coordination.md`",
			"`rules/testing-and-environment-validation.md`",
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
	// Agent entry files are verified against the rendered contract block; the
	// generated validation reference keeps its structural headings.
	return map[string][]string{
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
