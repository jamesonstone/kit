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
		"- Preserve the command's scope, phase, safety, and mutation boundaries.",
		"- The supervisor owns scope, integration, validation, evidence, delivery gates, and the final report.",
		"- When work separates into low-overlap areas, load `docs/references/rules/agent-team-orchestration.md`, predict touched files, and record an Agent Team Plan before spawning.",
		"- Use one lane for trivial, tightly coupled, ambiguous, or high-overlap work. In normal operation, run at most 3 independent lanes and serialize shared files or interfaces.",
		"- A fourth lane requires explicit exceptional authorization from the supervisor, clearly low file overlap, and an independent validation surface; never exceed 4 lanes.",
		"- After nontrivial implementation, use a read-only verification agent unless the task is documentation-only, tightly coupled, explicitly single-agent, or the runtime cannot spawn agents.",
		"- Subagents may use only a supervisor-prepared, explicitly assigned worktree; they may not create, switch, move, or remove worktrees, expand scope, or mutate Git/GitHub delivery state without explicit supervisor authorization.",
		"- Report actual agents used and omitted lanes. If none ran, state: `single supervisor lane; no specialist or verification agents spawned`.",
	}, "\n")
}
