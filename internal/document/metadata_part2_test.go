package document

import (
	"strings"
	"testing"
)

func TestValidateRejectsInvalidMetadataEnums(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
relationships:
  - type: follows
    target: 0002-beta
references:
  - name: Thing
    type: doc
    target: docs/thing.md
    selector_type: line
    selector: 12
    relation: adjacent
    read_policy: maybe
    used_for: context
    status: maybe
---
# SPEC

## SUMMARY

summary

## PROBLEM

problem

## GOALS

goals

## NON-GOALS

not applicable

## USERS

users

## SKILLS

skills are tracked in front matter.

## RELATIONSHIPS

none

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

requirements

## ACCEPTANCE

acceptance

## EDGE-CASES

not applicable

## OPEN-QUESTIONS

not required
`, "SPEC.md", TypeSpec)

	var relationshipError, referenceError bool
	for _, err := range doc.Validate() {
		if strings.Contains(err.Message, "invalid relationship type") {
			relationshipError = true
		}
		if strings.Contains(err.Message, "invalid reference relation") ||
			strings.Contains(err.Message, "invalid reference read_policy") ||
			strings.Contains(err.Message, "invalid reference selector_type") ||
			strings.Contains(err.Message, "invalid reference status") {
			referenceError = true
		}
	}
	if !relationshipError || !referenceError {
		t.Fatalf("Validate() relationship error = %v, reference error = %v, errors = %#v", relationshipError, referenceError, doc.Validate())
	}
}

func TestValidateWarnsForReferencePolicyMismatches(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
references:
  - name: Stale doc
    type: doc
    target: docs/stale.md
    relation: informs
    read_policy: conditional
    used_for: old context
    status: stale
  - name: Constraint
    type: doc
    target: docs/constraint.md
    relation: constrains
    read_policy: conditional
    used_for: constraints
    status: active
  - name: Evidence
    type: doc
    target: docs/evidence.md
    selector: Results
    relation: verifies
    read_policy: conditional
    used_for: verification
    status: active
---
# SPEC

## SUMMARY

summary

## PROBLEM

problem

## GOALS

goals

## NON-GOALS

not applicable

## USERS

users

## SKILLS

skills are tracked in front matter.

## RELATIONSHIPS

none

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

requirements

## ACCEPTANCE

acceptance

## EDGE-CASES

not applicable

## OPEN-QUESTIONS

not required
`, "SPEC.md", TypeSpec)

	var staleWarning, constraintWarning, evidenceWarning, selectorWarning bool
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity != MetadataDiagnosticWarning {
			continue
		}
		switch {
		case strings.Contains(diagnostic.Message, "stale reference should normally be skipped"):
			staleWarning = true
		case strings.Contains(diagnostic.Message, "constraining reference should normally be must-read"):
			constraintWarning = true
		case strings.Contains(diagnostic.Message, "verification reference should normally be evidence-read"):
			evidenceWarning = true
		case strings.Contains(diagnostic.Message, "reference selector is set without selector_type"):
			selectorWarning = true
		}
	}
	if !staleWarning || !constraintWarning || !evidenceWarning || !selectorWarning {
		t.Fatalf("warnings stale=%v constraint=%v evidence=%v selector=%v diagnostics=%#v", staleWarning, constraintWarning, evidenceWarning, selectorWarning, doc.MetadataDiagnostics)
	}
	if errors := doc.Validate(); len(errors) != 0 {
		t.Fatalf("Validate() errors = %#v, want warnings only", errors)
	}
}
