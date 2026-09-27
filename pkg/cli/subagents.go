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

func preparePromptWithoutSubagents(prompt string) string {
	return preparePrompt(prompt, false)
}

func preparePrompt(prompt string, includeSubagents bool) string {
	return preparePromptWithProfile(prompt, includeSubagents, currentPromptProfile())
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
		"- Parallelize independent investigation or implementation when the host supports it and it helps; keep small or tightly coupled work in the primary agent.",
		"- The primary agent owns scope, integration, validation, delivery, and the final report. Delegated agents never mutate Git or GitHub delivery state and never write overlapping files concurrently.",
		"- Follow `docs/references/rules/agent-team-orchestration.md` when delegating, and report only topology and verification that actually ran.",
	}, "\n")
}
