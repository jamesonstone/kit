package cli

import "strings"

var singleAgent bool
var legacySubagents bool

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&singleAgent,
		"single-agent",
		false,
		"disable default subagent orchestration guidance and keep prompts in one lane",
	)

	rootCmd.PersistentFlags().BoolVar(
		&legacySubagents,
		"subagents",
		false,
		"deprecated: subagents are now enabled by default",
	)
	if flag := rootCmd.PersistentFlags().Lookup("subagents"); flag != nil {
		flag.Hidden = true
		flag.Deprecated = "subagents are enabled by default; use --single-agent to disable orchestration guidance"
	}
}

func prepareAgentPrompt(prompt string) string {
	return preparePrompt(prompt, !singleAgent)
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
	return strings.Join([]string{
		"## Subagent Orchestration",
		"- preserve the command-specific rules above; this section adds routing guidance and does not relax phase, scope, or safety constraints",
		"- drive to understanding first: read the available context, identify uncertainties, and confirm scope before proposing or executing changes",
		"- then drive task orchestration coordination: turn the work into explicit workstreams and manage them centrally",
		"- use intelligent routing to identify the different areas of change or analysis and default to subagents when the work spans multiple distinct areas",
		"- before parallelizing, predict likely touched files or interfaces, cluster overlap conservatively, and avoid unsafe parallelism",
		"- when overlap or ambiguity is high, apply the same discovery-first discipline as kit dispatch and keep the queue conservative",
		"- assign one subagent per distinct, low-overlap area; keep overlapping or ambiguous work with the same subagent",
		"- parallelize only independent areas and serialize dependent or cross-cutting work",
		"- when a subagent needs an isolated checkout, use `git worktree add ~/worktrees/<repo>-<branch> <branch>` or `git worktree add -b <branch> ~/worktrees/<repo>-<branch> <base-ref>` and keep all worktrees flat under `~/worktrees/`",
		"- keep the main agent responsible for synthesis, final integration, validation, and communication",
	}, "\n")
}
