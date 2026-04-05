package cli

import "strings"

func findingsIssues(findings []reconcileFinding) string {
	var parts []string
	for _, finding := range findings {
		parts = append(parts, finding.Issue)
	}
	return strings.Join(parts, "\n")
}

func validConstitution() string {
	return `# CONSTITUTION

## PRINCIPLES

principles

## CONSTRAINTS

constraints

## CHANGE CLASSIFICATION

classification

## NON-GOALS

non-goals

## DEFINITIONS

definitions
`
}

func validProgressSummary(id, slug string) string {
	row := ""
	summary := "none"
	if id != "" && slug != "" {
		row = "| " + id + " | " + slug + " | `docs/specs/" + id + "-" + slug + "` | tasks | no | 2026-04-05 | summary |\n"
		summary = "### " + slug + "\n\n- **STATUS**: tasks\n- **PAUSED**: no\n- **INTENT**: summary\n- **APPROACH**: summary\n- **OPEN ITEMS**: none\n- **POINTERS**: `docs/specs/" + id + "-" + slug + "/SPEC.md`, `docs/specs/" + id + "-" + slug + "/PLAN.md`, `docs/specs/" + id + "-" + slug + "/TASKS.md`\n"
	}

	return "# PROJECT PROGRESS SUMMARY\n\n## FEATURE PROGRESS TABLE\n\n| ID | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |\n| -- | ------- | ---- | ----- | ------ | ------- | ------- |\n" + row + "\n## PROJECT INTENT\n\nintent\n\n## GLOBAL CONSTRAINTS\n\nconstraints\n\n## FEATURE SUMMARIES\n\n" + summary + "\n## LAST UPDATED\n\n2026-04-05 07:01:17 EDT\n"
}

func validSpecWithRelationships(relationships string) string {
	return `# SPEC

## SUMMARY

summary

## PROBLEM

problem

## GOALS

goals

## NON-GOALS

non-goals

## USERS

users

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

` + relationships + `
## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no supporting dependencies recorded yet | active |

## REQUIREMENTS

requirements

## ACCEPTANCE

acceptance

## EDGE-CASES

none

## OPEN-QUESTIONS

none
`
}

func validPlan() string {
	return `# PLAN

## SUMMARY

summary

## APPROACH

approach

## COMPONENTS

components

## DATA

data

## INTERFACES

interfaces

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no planning dependencies recorded yet | active |

## RISKS

risks

## TESTING

testing
`
}

func validTasks() string {
	return `# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | sample task | todo | agent | |

## TASK LIST

Use markdown checkboxes to track completion:

- [ ] T001: sample task

## TASK DETAILS

### T001
- **GOAL**: goal
- **SCOPE**: scope
- **ACCEPTANCE**: acceptance

## DEPENDENCIES

none

## NOTES

none
`
}

func invalidTasksMissingDetail() string {
	return `# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | sample task | todo | agent | |

## TASK LIST

Use markdown checkboxes to track completion:

- [ ] T001: sample task

## TASK DETAILS

### T002
- **GOAL**: goal
- **SCOPE**: scope
- **ACCEPTANCE**: acceptance

## DEPENDENCIES

none

## NOTES

none
`
}
