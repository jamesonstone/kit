package document

import (
	"strings"
	"testing"
)

func TestValidateV3SpecUsesPhaseAwareContentGates(t *testing.T) {
	ready := `---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: ready
feature:
  id: 0001
  slug: sample
  dir: 0001-sample
---
# SPEC

## PURPOSE

purpose

## CONTEXT

context

## REQUIREMENTS

requirements, non-goals, and observable acceptance

## ACCEPTED PLAN

accepted plan

## DECISIONS

<!-- TODO: pending implementation decisions -->

## DISCOVERIES

<!-- TODO: pending implementation discoveries -->

## VALIDATION

<!-- TODO: pending validation -->

## OUTCOME

<!-- TODO: pending outcome -->

## REPOSITORY MEMORY

<!-- TODO: pending curation -->
`
	if errors := Parse(ready, "SPEC.md", TypeSpec).Validate(); len(errors) != 0 {
		t.Fatalf("ready V3 validation errors = %#v, want none", errors)
	}

	deliver := Parse(strings.Replace(ready, "phase: ready", "phase: deliver", 1), "SPEC.md", TypeSpec)
	if errors := deliver.Validate(); len(errors) != 5 {
		t.Fatalf("deliver V3 validation errors = %d, want 5: %#v", len(errors), errors)
	}
}

func TestValidateV3SpecCompletionAcceptsResolvedSectionsWithoutACIdentifiers(t *testing.T) {
	content := `---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0001
  slug: sample
  dir: 0001-sample
---
# SPEC

## PURPOSE
purpose
## CONTEXT
context
## REQUIREMENTS
observable acceptance without identifiers
## ACCEPTED PLAN
plan
## DECISIONS
not required
## DISCOVERIES
not required
## VALIDATION
go test ./... passed
## OUTCOME
implemented
## REPOSITORY MEMORY
Decision: updated
`
	if errors := Parse(content, "SPEC.md", TypeSpec).Validate(); len(errors) != 0 {
		t.Fatalf("complete V3 validation errors = %#v, want none", errors)
	}
}

func TestV3SummaryFallsBackToPurpose(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: ready
feature:
  id: "0001"
  slug: sample
  dir: 0001-sample
---
# SPEC

## PURPOSE

Keep consequential implementation rationale in the repository.

## CONTEXT

Native planning is transient.
`, "SPEC.md", TypeSpec)

	if got := doc.SummaryText(); got != "Keep consequential implementation rationale in the repository." {
		t.Fatalf("SummaryText() = %q", got)
	}
}

func TestValidateSpecRequiresSummarySkillsAndDependencies(t *testing.T) {
	doc := Parse(`# SPEC

## PROBLEM

problem

## GOALS

goals

## NON-GOALS

non-goals

## USERS

users

## REQUIREMENTS

requirements

## ACCEPTANCE

acceptance

## EDGE-CASES

edge-cases

## OPEN-QUESTIONS

open-questions
`, "SPEC.md", TypeSpec)

	errors := doc.Validate()

	required := map[string]bool{
		"SUMMARY":       false,
		"SKILLS":        false,
		"RELATIONSHIPS": false,
		"DEPENDENCIES":  false,
	}

	for _, err := range errors {
		if _, ok := required[err.Section]; ok {
			required[err.Section] = true
		}
	}

	for section, found := range required {
		if !found {
			t.Fatalf("expected missing section %q to be reported, got %#v", section, errors)
		}
	}
}

func TestValidateBrainstormRequiresDependencies(t *testing.T) {
	doc := Parse(`# BRAINSTORM

## SUMMARY

summary

## USER THESIS

thesis

## CODEBASE FINDINGS

findings

## AFFECTED FILES

files

## QUESTIONS

questions

## OPTIONS

options

## RECOMMENDED STRATEGY

strategy

## NEXT STEP

next
`, "BRAINSTORM.md", TypeBrainstorm)

	errors := doc.Validate()

	for _, err := range errors {
		if err.Section == "DEPENDENCIES" || err.Section == "RELATIONSHIPS" {
			return
		}
	}

	t.Fatalf("expected missing DEPENDENCIES or RELATIONSHIPS section to be reported, got %#v", errors)
}
