// package templates provides embedded document templates for Kit.
package templates

import "strings"

// Constitution template per spec section 6.1
const Constitution = `# CONSTITUTION

## PRINCIPLES

<!-- TODO: define core principles that guide all decisions -->

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Spec-Driven (Formal)

<!-- use when: new features, kit brainstorm/kit spec, substantial architectural or behavioral changes -->
<!-- workflow: optional BRAINSTORM.md → SPEC.md → PLAN.md → TASKS.md → implement → reflect -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update only practical docs (READMEs, inline docs, API docs) -->
<!-- do NOT create SPEC.md / PLAN.md / TASKS.md for ad hoc work -->

### Ad Hoc with Existing Specs

<!-- if change touches code with existing spec docs: default to updating them -->
<!-- skip spec updates only for purely mechanical changes (formatting, typo, dep bump) -->

## NON-GOALS

<!-- TODO: define what this project explicitly will not do -->

## DEFINITIONS

<!-- TODO: define key terms used throughout the project -->
`

// BrainstormArtifact template for pre-spec research.
const BrainstormArtifact = `# BRAINSTORM

## SUMMARY

<!-- TODO: 1-2 sentence summary of the issue, opportunity, and likely direction -->

## USER THESIS

<!-- TODO: capture the user's issue or feature description in their own terms -->

## RELATIONSHIPS

none

## CODEBASE FINDINGS

<!-- TODO: summarize relevant architecture, patterns, constraints, and related flows -->

## AFFECTED FILES

<!-- TODO: list concrete file paths and why they matter -->

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |

<!-- TODO: list the tools, docs, design refs, APIs, libraries, datasets, assets, and other inputs used during this phase; keep exact URLs or file/node refs in Location -->

## QUESTIONS

<!-- TODO: list unresolved clarifying questions and unknowns -->

## OPTIONS

<!-- TODO: compare viable strategies and tradeoffs -->

## RECOMMENDED STRATEGY

<!-- TODO: document the preferred direction and why -->

## NEXT STEP

<!-- TODO: state the next workflow step, usually kit spec <feature> -->
`

// BuildBrainstormArtifact seeds a new brainstorm document with the user's thesis.
func BuildBrainstormArtifact(userThesis string) string {
	userThesis = strings.TrimSpace(userThesis)
	if userThesis == "" {
		return BrainstormArtifact
	}

	return strings.Replace(
		BrainstormArtifact,
		"<!-- TODO: capture the user's issue or feature description in their own terms -->",
		userThesis,
		1,
	)
}

// Spec template per spec section 6.2
const Spec = `# SPEC

## SUMMARY

<!-- TODO: 1-2 sentence business summary of this feature -->

## PROBLEM

<!-- TODO: describe the problem being solved -->

## GOALS

<!-- TODO: list what this feature must achieve -->

## NON-GOALS

<!-- TODO: list what this feature will not do -->

## USERS

<!-- TODO: identify who will use this feature -->

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no supporting dependencies recorded yet | active |

<!-- TODO: list the supporting docs, MCP tools, design refs, APIs, libraries, datasets, assets, and other inputs that shaped this spec -->

## REQUIREMENTS

<!-- TODO: list functional requirements -->

## ACCEPTANCE

<!-- TODO: define acceptance criteria -->

## EDGE-CASES

<!-- TODO: document edge cases and how they should be handled -->

## OPEN-QUESTIONS

<!-- TODO: list unresolved questions -->
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

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no planning dependencies recorded yet | active |

<!-- TODO: list the dependencies that shape the implementation strategy, including exact design URLs or file/node refs when applicable -->

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
- **NOTES**: <!-- only if necessary -->

## DEPENDENCIES

<!-- TODO: document task dependencies and ordering -->

## NOTES

<!-- TODO: additional context or implementation notes -->
`
