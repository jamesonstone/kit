package cli

import "strings"

var subagents bool

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&subagents,
		"subagents",
		false,
		"append discovery-first subagent orchestration guidance to agent prompts",
	)
}

func prepareAgentPrompt(prompt string) string {
	return preparePrompt(prompt, subagents)
}

func preparePromptWithoutSubagents(prompt string) string {
	return preparePrompt(prompt, false)
}

func preparePrompt(prompt string, includeSubagents bool) string {
	prompt = appendSkillPromptSuffix(prompt)

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
	return strings.TrimLeft(`
## Subagent Orchestration
- preserve the command-specific rules above; this section adds routing guidance and does not relax phase, scope, or safety constraints
- drive to understanding first: read the available context, identify uncertainties, and confirm scope before proposing or executing changes
- then drive task orchestration coordination: turn the work into explicit workstreams and manage them centrally
- use intelligent routing to identify the different areas of change or analysis
- when the work spans multiple areas, apply the same discovery-first discipline as kit dispatch: predict likely touched files or interfaces, cluster overlap conservatively, and avoid unsafe parallelism
- delegate and dispatch to subagents where possible
- assign one subagent per distinct, low-overlap area; keep overlapping or ambiguous work with the same subagent
- parallelize only independent areas and serialize dependent or cross-cutting work
- keep the main agent responsible for synthesis, final integration, validation, and communication
`, "\n")
}
