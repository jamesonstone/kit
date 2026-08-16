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
		"- The supervisor owns scope, integration, validation, evidence, delivery gates, and the final report.",
		"- Before assigning lanes, enter `CAPABILITY_NEGOTIATING` and build a Capability Manifest from only the agent controls exposed by the active runtime. Record host-confirmed child execution, parallelism, stable references and follow-up, model and effort selection, fresh verification, waiting or status controls, effective capacity, selected topology, delegation depth, degradations, and evidence basis. Preserve `unknown` literally, treat it as unavailable for routing, and never spawn only to probe capability.",
		"- Load `docs/references/rules/agent-team-orchestration.md`, predict touched files, and record an Agent Team Plan. Use one supervisor lane for trivial, tightly coupled, ambiguous, high-overlap, explicitly single-agent work, or when no child primitive is confirmed. Only the root supervisor may launch agents; child agents must not delegate further.",
		"- Count a lane as an actual agent only when the runtime creates a separate execution and returns a separate result. Report logical-only and omitted lanes separately; role prompts, task lists, editor modes, handoffs, and manually opened conversations are not spawned agents.",
		"- Record requested and effective provider-neutral profiles separately. If the runtime does not confirm model or effort, report `runtime-selected/unverified`; claim parallel execution only when the host confirms overlap, otherwise report sequential or unconfirmed execution.",
		"- Reuse the same stable agent reference for follow-up when supported. Otherwise fully rebrief a replacement and report continuity loss; never describe a replacement as the same agent.",
		"- After nontrivial implementation, use a fresh independent read-only verifier when supported. Otherwise perform a distinct read-only supervisor self-review and state that verification was not independent.",
		"- Subagents may use only a supervisor-prepared, explicitly assigned worktree; they may not create, switch, move, or remove worktrees, expand scope, or mutate Git/GitHub delivery state without explicit supervisor authorization.",
		"- Final reporting must distinguish task outcome from orchestration conformance and list actual agents, logical lanes, omitted lanes, requested and effective profiles, confirmed or unconfirmed parallelism, continuity replacements, and `verification_independent: true | false | unknown`. If none ran, state: `single supervisor lane; no specialist or verification agents spawned`.",
	}, "\n")
}
