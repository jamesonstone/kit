// package templates provides embedded document templates for Kit.
package templates

import (
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

// Gitignore is the default Kit-local ignore block for repositories initialized
// with Kit. It intentionally does not ignore all of .kit/ so future tracked
// schema, README, or fixture files remain possible.
const Gitignore = `# Kit local generated environment, cache, and scratch artifacts
.env
.envrc
.kit/runs/
.kit/loops/
.kit/state.json
.kit/cache/
.kit/tmp/
.kit/temp/
.kit/*.tmp
.kit/*.lock
`

const Envrc = `#!/bin/sh
set -eu

dotenv_if_exists
`

// Constitution template per spec section 6.1
const Constitution = `# CONSTITUTION

## PRINCIPLES

<!-- TODO: define core principles that guide all decisions -->

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat ` + "`docs/CONSTITUTION.md`" + ` as the canonical project contract.
- Keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all ` + "`docs/**`" + `, all ` + "`.kit/**`" + `, or ` + "`.kit.yaml`" + `.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->

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

References are tracked in front matter.

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

// BuildBrainstormArtifactForFeature seeds a new brainstorm document with typed
// front matter for the feature-specific metadata Kit can know at creation time.
func BuildBrainstormArtifactForFeature(userThesis string, feature document.FeatureMetadata, references []document.MetadataReference) string {
	content := BuildBrainstormArtifact(userThesis)
	content = replaceTemplateSection(content, "RELATIONSHIPS", "Relationships are tracked in front matter.")
	content = replaceTemplateSection(content, "DEPENDENCIES", "References are tracked in front matter.")
	updated, _, err := document.UpsertMetadata(content, document.TypeBrainstorm, document.MetadataUpsert{
		Feature:    feature,
		References: references,
	})
	if err != nil {
		return content
	}
	return updated
}

func BuildSpecArtifactForFeature(feature document.FeatureMetadata) string {
	content := replaceTemplateSection(Spec, "SKILLS", "Skills are tracked in front matter.")
	content = replaceTemplateSection(content, "RELATIONSHIPS", "Relationships are tracked in front matter.")
	content = replaceTemplateSection(content, "DEPENDENCIES", "References are tracked in front matter.")
	updated, _, err := document.UpsertMetadata(content, document.TypeSpec, document.MetadataUpsert{
		Feature: feature,
	})
	if err != nil {
		return content
	}
	return updated
}

func BuildPlanArtifactForFeature(feature document.FeatureMetadata) string {
	content := replaceTemplateSection(Plan, "DEPENDENCIES", "References are tracked in front matter.")
	updated, _, err := document.UpsertMetadata(content, document.TypePlan, document.MetadataUpsert{
		Feature: feature,
	})
	if err != nil {
		return content
	}
	return updated
}

func BuildTasksArtifactForFeature(feature document.FeatureMetadata) string {
	updated, _, err := document.UpsertMetadata(Tasks, document.TypeTasks, document.MetadataUpsert{
		Feature: feature,
	})
	if err != nil {
		return Tasks
	}
	return updated
}

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

References are tracked in front matter.

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
