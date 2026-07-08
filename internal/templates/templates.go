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
- Treat ` + "`docs/notes/<feature>`" + ` as optional source material, not canonical truth; promote durable decisions into ` + "`SPEC.md`" + `, ` + "`docs/CONSTITUTION.md`" + `, or durable references.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all ` + "`docs/**`" + `, all ` + "`.kit/**`" + `, or ` + "`.kit.yaml`" + `.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Spec-Driven (Formal)

<!-- use when: new features, kit spec, substantial architectural or behavioral changes -->
<!-- workflow: kit spec <feature> → SPEC.md phases: clarify → ready → implement → validate → reflect → deliver -->
<!-- legacy staged documents: BRAINSTORM.md, legacy SPEC.md, PLAN.md, TASKS.md only when explicitly chosen -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update only practical docs (READMEs, inline docs, API docs) -->
<!-- do NOT create feature SPEC.md or legacy staged artifacts for ad hoc work -->

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
		Feature:         feature,
		WorkflowVersion: 2,
		Phase:           "clarify",
		Clarification:   clarificationMetadata(document.ClarificationStatusOpen, 0, 1),
	})
	if err != nil {
		return content
	}
	return updated
}

func clarificationMetadata(status string, confidence int, unresolvedQuestions int) *document.MetadataClarification {
	clarification := document.NewMetadataClarification(status, confidence, unresolvedQuestions)
	return &clarification
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
