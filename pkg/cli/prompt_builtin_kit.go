package cli

import "github.com/jamesonstone/kit/internal/promptlib"

func builtInPromptSources() []promptlib.Source {
	return []promptlib.Source{
		builtInToolboxPromptSource(),
		builtInKitPromptSource(),
	}
}

func builtInKitPromptSource() promptlib.Source {
	return promptlib.Source{
		Kind:     promptlib.SourceBuiltin,
		Location: builtInPromptLocation,
		Prompts: []promptlib.Prompt{
			dynamicBuiltInPrompt("workflow", "brainstorm", "Regenerate the brainstorm prompt for an existing feature.", []string{promptContextActiveFeature}, renderWorkflowBrainstormPrompt),
			dynamicBuiltInPrompt("workflow", "spec", "Regenerate the specification prompt for an existing feature.", []string{promptContextActiveFeature}, renderWorkflowSpecPrompt),
			dynamicBuiltInPrompt("workflow", "plan", "Regenerate the implementation-plan prompt for an existing feature.", []string{promptContextActiveFeature}, renderWorkflowPlanPrompt),
			dynamicBuiltInPrompt("workflow", "tasks", "Regenerate the task-planning prompt for an existing feature.", []string{promptContextActiveFeature}, renderWorkflowTasksPrompt),
			dynamicBuiltInPrompt("workflow", "implement", "Generate the implementation task-loop prompt for an existing feature.", []string{promptContextActiveFeature}, renderWorkflowImplementPrompt),
			dynamicBuiltInPrompt("workflow", "reflect", "Generate the reflection and verification prompt for current changes.", []string{promptContextOptionalFeature}, renderWorkflowReflectPrompt),
			dynamicBuiltInPrompt("support", "resume", "Generate a catch-up prompt for resuming an active feature.", []string{promptContextActiveFeature}, renderSupportResumePrompt),
			builtInPrompt("support", "handoff", "Generate documentation-sync handoff instructions.", renderSupportHandoffPrompt),
			builtInPrompt("support", "summarize", "Generate context-window summarization instructions.", renderSupportSummarizePrompt),
			dynamicBuiltInPrompt("support", "reconcile", "Generate a documentation reconciliation prompt.", []string{promptContextReconciliationReport}, renderSupportReconcilePrompt),
			dynamicBuiltInPrompt("support", "dispatch", "Generate a subagent dispatch-planning prompt.", []string{promptContextTaskList}, renderSupportDispatchPrompt),
			builtInPrompt("support", "code-review", "Generate branch code-review instructions.", renderSupportCodeReviewPrompt),
			dynamicBuiltInPrompt("skill", "mine", "Generate a reusable-skill mining prompt for a feature.", []string{promptContextActiveFeature}, renderSkillMinePrompt),
			builtInPrompt("project", "init", "Generate project constitution drafting instructions.", renderProjectInitPrompt),
			dynamicBuiltInPrompt("project", "refresh", "Generate project-level documentation refresh instructions.", []string{promptContextProject}, renderProjectRefreshPrompt),
		},
	}
}

func builtInPrompt(noun, verb, description string, render promptlib.RenderFunc) promptlib.Prompt {
	return promptlib.Prompt{
		Identity:    promptlib.Identity{Noun: noun, Verb: verb},
		Description: description,
		Render:      render,
	}
}

func dynamicBuiltInPrompt(
	noun string,
	verb string,
	description string,
	requirements []string,
	render promptlib.RenderFunc,
) promptlib.Prompt {
	prompt := builtInPrompt(noun, verb, description, render)
	prompt.ContextRequirements = requirements
	return prompt
}
