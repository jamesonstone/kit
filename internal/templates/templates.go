// package templates provides embedded document templates for Kit.
package templates

// Constitution template per spec section 6.1
const Constitution = `# CONSTITUTION

## PRINCIPLES

<!-- TODO: define core principles that guide all decisions -->

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

## NON-GOALS

<!-- TODO: define what this project explicitly will not do -->

## DEFINITIONS

<!-- TODO: define key terms used throughout the project -->
`

// Spec template per spec section 6.2
const Spec = `# SPEC

## PROBLEM

<!-- TODO: describe the problem being solved -->

## GOALS

<!-- TODO: list what this feature must achieve -->

## NON-GOALS

<!-- TODO: list what this feature will not do -->

## USERS

<!-- TODO: identify who will use this feature -->

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

## RISKS

<!-- TODO: identify risks and mitigation strategies -->

## TESTING

<!-- TODO: describe testing strategy -->
`

// Tasks template per spec section 6.4
const Tasks = `# TASKS

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T-01 | <!-- TODO: task description --> | pending | <!-- owner --> | <!-- deps --> |

## TASKS

<!-- TODO: detailed task descriptions linking to [PLAN-XX] items -->

## DEPENDENCIES

<!-- TODO: document task dependencies and ordering -->

## NOTES

<!-- TODO: additional context or implementation notes -->
`

// Analysis template per spec section 6.6
const Analysis = `# ANALYSIS

## UNDERSTANDING

**Current Understanding: 0%%**

<!-- understanding percentage tracked at top and bottom -->

## QUESTIONS

<!-- TODO: open questions for the user/team -->

## RESEARCH

<!-- technical investigation notes: library comparisons, performance benchmarks, compatibility findings -->

## CLARIFICATIONS

<!-- resolved questions with answers -->

## ASSUMPTIONS

<!-- documented assumptions made during analysis -->

## RISKS

<!-- identified risks or concerns -->

---

**Understanding: 0%%**
`

// ProjectProgressSummary template per spec section 6.5
const ProjectProgressSummary = `# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------- | ------- |

## PROJECT INTENT

<!-- TODO: describe the overall project purpose -->

## GLOBAL CONSTRAINTS

<!-- TODO: summarize key constraints from CONSTITUTION.md -->

## FEATURE SUMMARIES

<!-- feature summaries will be generated here -->

## LAST UPDATED

<!-- timestamp updated by kit rollup -->
`

// AgentPointer returns a template for agent pointer files.
func AgentPointer(agentName string) string {
	return `# ` + agentName + `

## Kit is the source of truth

- Constitution: ` + "`docs/CONSTITUTION.md`" + `
- Feature specs live under: ` + "`docs/specs/<feature>/`" + `
  - ` + "`SPEC.md`" + ` (requirements)
  - ` + "`PLAN.md`" + ` (implementation plan)
  - ` + "`TASKS.md`" + ` (executable task list)
  - ` + "`ANALYSIS.md`" + ` (optional, analysis scratchpad)

## Workflow contract

- Specs drive code. Code serves specs.
- For any change:
  1. locate the relevant feature directory in ` + "`docs/specs/<feature>/`" + `
  2. read ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + `
  3. implement tasks in order
  4. verify (tests / validation steps from plan)
  5. if reality diverges, update ` + "`SPEC.md`" + ` / ` + "`PLAN.md`" + ` / ` + "`TASKS.md`" + ` first, then code

## Multi-feature rule

- Never mix features in one ` + "`docs/specs/<feature>/`" + ` directory.
- If work spans features, update each feature's docs separately.
`
}

// FeatureSummaryTemplate returns a template for a feature summary in PROJECT_PROGRESS_SUMMARY.md
const FeatureSummaryTemplate = `### {{.FeatureName}}

- **STATUS**: {{.Phase}}
- **INTENT**: {{.Intent}}
- **APPROACH**: {{.Approach}}
- **OPEN ITEMS**: {{.OpenItems}}
- **POINTERS**: ` + "`{{.Path}}/SPEC.md`" + `, ` + "`{{.Path}}/PLAN.md`" + `, ` + "`{{.Path}}/TASKS.md`" + `
`
