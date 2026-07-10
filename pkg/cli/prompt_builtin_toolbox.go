package cli

import "github.com/jamesonstone/kit/internal/promptlib"

const builtInPromptLocation = "built-in"

func builtInToolboxPromptSource() promptlib.Source {
	return promptlib.Source{
		Kind:     promptlib.SourceBuiltin,
		Location: builtInPromptLocation,
		Prompts: []promptlib.Prompt{
			{
				Identity:    promptlib.Identity{Noun: "coding-agent", Verb: "short"},
				Description: "Short clarification-before-implementation prompt.",
				Content:     codingAgentShortPrompt,
			},
			{
				Identity:    promptlib.Identity{Noun: "coding-agent", Verb: "long"},
				Description: "Detailed clarification-before-implementation prompt.",
				Content:     codingAgentLongPrompt,
			},
			{
				Identity:    promptlib.Identity{Noun: "coding-agent", Verb: "instructions"},
				Description: "Implementation-ready coding agent instruction prompt.",
				Content:     codingAgentInstructionsPrompt,
			},
		},
	}
}

const codingAgentShortPrompt = "Clarify only material choices that repository or task evidence cannot resolve. Do not implement yet. Ask concise numbered questions with recommended defaults and impact; otherwise state that no input is needed."

const codingAgentLongPrompt = `Stay in explicit clarification workflow. Do not implement or make production changes.

First inspect the available task and repository evidence. Resolve discoverable facts yourself. Ask only about a remaining choice that materially changes scope, behavior, risk, validation, or delivery.

For each numbered question include:
- the decision needed;
- a recommended default;
- why the answer changes the result;
- a compact answer format.

If no material question remains, output "Open Questions: none" and a concise implementation-ready summary of goal, constraints, success criteria, and known risks. Record durable decisions in the active spec or planning artifact when authorized.

Stop after the questions when user input is required; do not append implementation instructions.`

const codingAgentInstructionsPrompt = `Output implementation-ready coding-agent instructions in one markdown code block and no surrounding commentary.

Use only sections that materially apply, while covering:

## Goal
- The user outcome and why it matters.

## Context
- Relevant repository state, existing patterns, and source-of-truth paths.

## Scope And Constraints
- In-scope and out-of-scope behavior, approval boundaries, compatibility, security, performance, and reliability constraints.

## Requirements And Design
- Functional behavior, interfaces/data contracts, failure modes, and the simplest viable approach.

## Implementation
- Ordered coherent steps and exact likely files when known; use repository discovery instead of invented paths.

## Validation
- Focused tests, integration/runtime/manual checks, documentation evidence, and what each proves.

## Acceptance Criteria
- Binary definition of done, including edge cases and required output.

Do not manufacture missing facts. Direct the agent to discover safe repository facts, and reserve questions for material non-discoverable choices. State each instruction once.`
