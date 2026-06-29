package cli

import "strings"

var singleAgent bool

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&singleAgent,
		"single-agent",
		false,
		"disable default subagent orchestration guidance and keep prompts in one lane",
	)
}

func prepareAgentPrompt(prompt string) string {
	return preparePrompt(prompt, !singleAgent)
}

func prepareAgentPromptForFeature(prompt, featurePath string) string {
	return preparePromptForFeature(prompt, !singleAgent, featurePath)
}

func preparePromptWithoutSubagents(prompt string) string {
	return preparePrompt(prompt, false)
}

func preparePrompt(prompt string, includeSubagents bool) string {
	return preparePromptWithProfile(prompt, includeSubagents, currentPromptProfile())
}

func preparePromptForFeature(prompt string, includeSubagents bool, featurePath string) string {
	return preparePromptWithProfile(prompt, includeSubagents, effectivePromptProfile(featurePath))
}

func preparePromptWithProfile(prompt string, includeSubagents bool, profile promptProfile) string {
	prompt = appendSkillPromptSuffix(prompt)
	prompt = appendPromptProfileSuffix(prompt, profile)

	if !includeSubagents {
		return prompt
	}

	trimmedPrompt := strings.TrimRight(prompt, "\n")
	if trimmedPrompt == "" {
		return subagentPromptSuffix()
	}

	return trimmedPrompt + "\n\n" + subagentPromptSuffix()
}

func subagentPromptSuffix() string {
	return strings.Join([]string{
		"## Subagent Orchestration",
		"- preserve the command-specific rules above; this section adds routing guidance and does not relax phase, scope, or safety constraints",
		"- when command-specific rules limit the phase to documentation, execution means documentation edits only; do not modify product code, tests, runtime config, generated artifacts, or implementation files",
		"- drive to understanding first: read the available context, identify uncertainties, and confirm scope before proposing or executing changes",
		"- when full-context loading would be noisy or wasteful, do RLM-style discovery first: identify the immediate decision, load the smallest relevant artifact, and stop once the decision is supported",
		"- when execution topology matters and `docs/references/rules/agent-team-orchestration.md` exists, load it before deciding whether to use specialist or verification subagents",
		"- act as one accountable supervisor: own scope, lane assignment, integration, validation, evidence, delivery gating, and final reporting; do not delegate final responsibility",
		"- before implementation, prepare an Agent Team Plan with supervisor responsibilities, proposed lanes, subagents actually spawned, logical-only lanes, omitted lanes with reasons, predicted touched files, overlap risks, max concurrency, serialized work, and validation/review lanes",
		"- use subagents only when low-overlap lanes improve correctness or throughput; do not use subagents merely because they are available",
		"- keep work in a single supervisor lane when the task is trivial, tightly coupled, high-overlap, high-ambiguity, requires continuous design judgment, the runtime cannot spawn subagents, or the user requested single-agent execution; record the reason",
		"- default maximum concurrent lanes is 3; hard ceiling is 4; use 4 only when predicted file overlap is clearly low and each lane has independent validation",
		"- do not turn broad discovery into parallel execution until the scope is narrow enough to predict overlap with reasonable confidence",
		"- before parallelizing, predict likely touched files or interfaces, cluster overlap conservatively, and avoid unsafe parallelism",
		"- when overlap or ambiguity is high, apply the same discovery-first discipline as kit dispatch and keep the queue conservative",
		"- assign one implementation subagent per distinct, low-overlap area; keep overlapping or ambiguous work with the supervisor or one serialized lane",
		"- after implementation, use at least one read-only verification subagent by default unless the change is documentation-only, trivial, tightly coupled, the runtime cannot spawn subagents, or the user requested single-agent execution",
		"- verification agents are read-only by default: they must not edit files, stage changes, commit, push, resolve review threads, or mark acceptance criteria complete",
		"- subagents must not independently expand scope or mutate delivery state unless explicitly assigned and allowed by the supervisor",
		"- parallelize only independent areas and serialize dependent or cross-cutting work",
		"- keep all subagent work in the existing project directory; do not create or use git worktrees",
		"- if the current branch or dirty state is unsuitable for a subagent, stop and ask the user how to proceed instead of creating an alternate checkout",
		"- final responses must state actual subagents spawned, logical lanes not spawned, and any single-lane exception; if no separate agents ran, state exactly: `single supervisor lane; no specialist or verification agents spawned`",
	}, "\n")
}
