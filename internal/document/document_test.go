package document

import "testing"

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
		"SUMMARY":      false,
		"SKILLS":       false,
		"DEPENDENCIES": false,
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
		if err.Section == "DEPENDENCIES" {
			return
		}
	}

	t.Fatalf("expected missing DEPENDENCIES section to be reported, got %#v", errors)
}

func TestValidatePlanRequiresDependencies(t *testing.T) {
	doc := Parse(`# PLAN

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

## RISKS

risks

## TESTING

testing
`, "PLAN.md", TypePlan)

	errors := doc.Validate()

	for _, err := range errors {
		if err.Section == "DEPENDENCIES" {
			return
		}
	}

	t.Fatalf("expected missing DEPENDENCIES section to be reported, got %#v", errors)
}
