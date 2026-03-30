package document

import "testing"

func TestValidateSpecRequiresSummaryAndSkills(t *testing.T) {
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
		"SUMMARY": false,
		"SKILLS":  false,
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
