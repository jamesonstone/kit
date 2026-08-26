// package templates provides embedded document templates for Kit.
package templates

import (
	"strings"

	"github.com/jamesonstone/kit/v3/internal/document"
)

// Gitignore is the default Kit-local ignore block for repositories initialized
// with Kit. It intentionally does not ignore all of .kit/ so future tracked
// schema, README, or fixture files remain possible.
const Gitignore = `# Kit local generated environment, cache, and scratch artifacts
.env
.envrc
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

// Makefile is the safe starter command surface for repositories initialized
// with Kit. Project-specific targets are added by the initialization prompt
// only after their underlying commands have been verified.
const Makefile = `.DEFAULT_GOAL := help

.PHONY: help

help:
	@printf '%s\n' 'Project developer workflow'
	@printf '%s\n' ''
	@printf '%s\n' 'Run the Kit initialization prompt to add project-specific targets.'
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
- Use native agent planning for research, clarification, design, and implementation planning.
- Before implementation, inspect code and repository memory; create or adopt ` + "`SPEC.md`" + ` when material rationale exists.
- After validation, curate feature rationale, project invariants, reusable practices, and domain knowledge into their scope-appropriate canonical documents.
- Allow a justified ` + "`not required`" + ` repository-memory decision when code and tests preserve the complete durable truth.
- Before a substantial terminal completion or handoff response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` and report only What happened, Deviations, and Next steps; answer ordinary conversational requests naturally without that structured envelope.
- Before commit, pull request, issue, comment, or other attribution text, load ` + "`docs/references/rules/human-authorship.md`" + `. Only the human user may be displayed as author; do not attribute coding agents, tools, or bots.
- Before designing deletion behavior or deleting persistent project, user, business, or external-system state, load ` + "`docs/references/rules/deletion-safety.md`" + `.
- Default unqualified deletion to a recoverable soft delete; require a post-outline specific manual confirmation for the exact current targets before any hard delete.
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less.
- Before delivery, audit the complete affected source/test scope; whole-project reconcile and scheduled maintenance audit the entire repository.
- Exclude documentation files, all ` + "`docs/**`" + `, all ` + "`.kit/**`" + `, ` + "`.kit.yaml`" + `, ignored files, vendored dependencies, and proven generated files.
- Split oversized files by semantic responsibility while preserving stable public entry points and behavior; never use minification or arbitrary numbered chunks to claim compliance.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Repository-Memory Work

<!-- use when: consequential product rationale, architecture, cross-component behavior, or historical decisions must survive -->
<!-- workflow: native plan → create/adopt SPEC.md before code → implement → validate → curate repository memory -->
<!-- legacy staged documents: BRAINSTORM.md, legacy SPEC.md, PLAN.md, TASKS.md only when explicitly chosen -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update practical canonical docs when behavior changes -->
<!-- do not create feature SPEC.md solely for ceremony; report a justified not-required memory decision -->

### Ad Hoc with Existing Specs

<!-- if change touches code with existing spec docs: update them when rationale, behavior, requirements, or approach changes -->
<!-- leave them unchanged when code and tests communicate the complete durable truth -->

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
	content := Spec
	updated, _, err := document.UpsertMetadata(content, document.TypeSpec, document.MetadataUpsert{
		Feature:         feature,
		WorkflowVersion: document.WorkflowVersionV3,
		Phase:           "clarify",
	})
	if err != nil {
		return content
	}
	return updated
}

func BuildSpecV2ArtifactForFeature(feature document.FeatureMetadata) string {
	content := SpecV2
	updated, _, err := document.UpsertMetadata(content, document.TypeSpec, document.MetadataUpsert{
		Feature:         feature,
		WorkflowVersion: document.WorkflowVersionV2,
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
