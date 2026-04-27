package templates

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

| ID | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------ | ------- | ------- |

## PROJECT INTENT

Kit is a document-first workflow harness for disciplined thought work. It keeps durable project context in canonical markdown artifacts so humans and coding agents can move from research to specification, planning, tasks, implementation, reflection, and completion with explicit traceability.

## GLOBAL CONSTRAINTS

See docs/CONSTITUTION.md for project-wide constraints and principles.

## FEATURE SUMMARIES

<!-- feature summaries will be generated here -->

## LAST UPDATED

<!-- timestamp updated by kit rollup -->
`

// FeatureSummaryTemplate returns a template for a feature summary in PROJECT_PROGRESS_SUMMARY.md
const FeatureSummaryTemplate = `### {{.FeatureName}}

- **STATUS**: {{.Phase}}
- **PAUSED**: {{.Paused}}
- **INTENT**: {{.Intent}}
- **APPROACH**: {{.Approach}}
- **OPEN ITEMS**: {{.OpenItems}}
- **POINTERS**: ` + "`{{.Path}}/SPEC.md`" + `, ` + "`{{.Path}}/PLAN.md`" + `, ` + "`{{.Path}}/TASKS.md`" + `
`
