package templates

import "strings"

func replaceTemplateSection(content, sectionName, sectionBody string) string {
	lines := strings.Split(content, "\n")
	header := "## " + sectionName
	start := -1
	end := len(lines)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if start == -1 {
			if trimmed == header {
				start = i
			}
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			end = i
			break
		}
	}
	if start == -1 {
		return content
	}

	replacementLines := []string{header, "", sectionBody, ""}
	updatedLines := append([]string{}, lines[:start]...)
	updatedLines = append(updatedLines, replacementLines...)
	updatedLines = append(updatedLines, lines[end:]...)
	return strings.Join(updatedLines, "\n")
}

// Spec template for the v2 single-artifact workflow.
const Spec = `# SPEC

## THESIS

<!-- TODO: capture the original idea, problem statement, or feature thesis in the user's terms -->

## CONTEXT

<!-- TODO: record repo-grounded research findings, relevant files, references, relationships, and constraints -->

## CLARIFICATIONS

<!-- TODO: record clarification questions, answers, accepted defaults, rejected assumptions, and current confidence -->

## REQUIREMENTS

<!-- TODO: list clarified requirements and non-goals with stable identifiers when useful -->

## ASSUMPTIONS

<!-- TODO: list accepted assumptions, removed assumptions, and any assumption that still blocks progress -->

## ACCEPTANCE CRITERIA

<!-- TODO: define binary-verifiable acceptance criteria that can be mapped 1:1 to validation evidence -->

## IMPLEMENTATION PLAN

<!-- TODO: document the planned implementation approach, touched areas, risks, and rollback strategy -->

## TASK CHECKLIST

<!-- TODO: keep a concise durable checklist mapping tasks to lanes, acceptance criteria, status, and evidence -->

## VALIDATION MAP

<!-- TODO: map each acceptance criterion to exact tests, checks, runtime proof, documentation review, and evidence links -->

## REFLECTION NOTES

<!-- TODO: record post-implementation review findings, fixes, remaining risks, and confidence -->

## DOCUMENTATION UPDATES

<!-- TODO: list affected docs and whether each has been updated, verified, or intentionally left unchanged -->

## DELIVERY DECISION

<!-- TODO: record delivery intent, delivery lane, issue/branch/PR decision, and delivery hard-gate status -->

## EVIDENCE

<!-- TODO: summarize validation evidence and link detailed logs or run artifacts such as .kit/runs entries -->
`

// Plan template per spec section 6.3
const Plan = `# PLAN

## SUMMARY

<!-- TODO: brief overview of the implementation approach -->

## APPROACH

<!-- TODO: explain the strategy, not code -->

## COMPONENTS

<!-- TODO: list major components and their responsibilities -->

## DATA

<!-- TODO: describe data structures and storage -->

## INTERFACES

<!-- TODO: define APIs, contracts, and integration points -->

## DEPENDENCIES

References are tracked in front matter.

## RISKS

<!-- TODO: identify risks and mitigation strategies -->

## TESTING

<!-- TODO: describe testing strategy -->
`

// Tasks template per spec section 6.4
// IMPORTANT: tasks use markdown checkboxes for progress tracking:
//   - [ ] incomplete task
//   - [x] completed task
const Tasks = `# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | <!-- task description --> | todo | <!-- owner --> | <!-- deps --> |

## TASK LIST

Use markdown checkboxes to track completion:

- [ ] T001: <!-- task description -->

## TASK DETAILS

For each task, provide:

### T001
- **GOAL**: <!-- one sentence outcome -->
- **SCOPE**: <!-- tight bullets, no fluff -->
- **ACCEPTANCE**: <!-- concrete checks -->
- **VERIFY**:
  - <!-- runnable command, for example go test ./... -->
- **EXPECTED FILES**:
  - <!-- paths expected to change -->
- **RISK**: <!-- Low/Medium/High plus short reason -->
- **ROLLBACK**: <!-- how to revert safely, or not required -->
- **NOTES**: <!-- only if necessary -->

## DEPENDENCIES

<!-- TODO: document task dependencies and ordering -->

## NOTES

<!-- TODO: additional context or implementation notes -->
`
