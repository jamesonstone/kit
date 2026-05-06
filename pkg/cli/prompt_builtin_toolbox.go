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
				Description: "Short planning-mode clarification prompt.",
				Content:     codingAgentShortPrompt,
			},
			{
				Identity:    promptlib.Identity{Noun: "coding-agent", Verb: "long"},
				Description: "Detailed planning-mode clarification prompt.",
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

const codingAgentShortPrompt = "Clarify before implementing. Stay in planning mode. Ask numbered questions with defaults, assumptions, and uncertainty until >=95% confidence and 0 unresolved. Accept y/n shorthand. Report confidence each batch."

const codingAgentLongPrompt = `Stay in planning and information-gathering mode. Do not implement code or make production changes yet.

Ask clarifying questions until you have >=95% confidence in the problem, requirements, constraints, edge cases, and solution approach. Do not assume missing details.

Ask questions in numbered batches of up to 10. For each question, include: the question, recommended default answer, current assumption or proposed solution, and remaining uncertainty.

After each batch, output:
Current confidence: X%
Unresolved assumptions: N
Waiting for approval, overrides, or corrections.

Accept shorthand: y approves all defaults; y 1,2 approves only those defaults; n rejects all defaults and requires replacements; n 3: <answer> rejects question 3 with replacement; mixed forms like y: 1,2; n: 3 - <reason> are allowed.

If only specific question numbers are approved, treat all others as unresolved. Continue asking batches until confidence is >=95%, unresolved assumptions are 0, durable decisions are captured in the active spec/brainstorm artifact, and the next implementation step is clear.`

const codingAgentInstructionsPrompt = `Output a concise, comprehensive set of coding agent instructions.

Formatting:
- Output MUST be a single markdown code block.
- Do NOT include any text outside the code block.
- Use clear section headers.

Goal:
- Produce implementation-ready instructions that allow a coding agent to build the feature end to end without additional clarification.

Required Sections:
1. Objective
   - What is being built and why.

2. Scope
   - In-scope vs out-of-scope.

3. Assumptions
   - Environment, dependencies, constraints.

4. Requirements
   - Functional requirements.
   - Non-functional requirements (performance, security, reliability).

5. Architecture / Design
   - High-level approach.
   - Key components and interactions.

6. Data Models / Schemas
   - Structures, types, contracts.

7. APIs / Interfaces
   - Endpoints, inputs, outputs, error cases.

8. Files to Create / Modify
   - Exact file paths.
   - Purpose of each file.

9. Implementation Steps
   - Ordered, atomic steps.
   - No gaps between steps.

10. Edge Cases
   - Failure modes and handling.

11. Validation
   - How to verify correctness manually.

12. Tests
   - Unit tests.
   - Integration tests.
   - Include key test cases.

13. Acceptance Criteria
   - Clear, testable definition of “done”.

Constraints:
- Be specific. Avoid vague language.
- Prefer explicit examples over abstract descriptions.
- Minimize assumptions; state them if required.
- Optimize for correctness and completeness over brevity.

Output Rules:
- No commentary.
- No explanations outside instructions.
- No placeholders like “TODO” unless explicitly required.`
