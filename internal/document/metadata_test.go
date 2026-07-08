package document

import (
	"testing"
)

func TestParseFrontMatterKeepsBodySections(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
summary: Metadata summary
---
# SPEC

## SUMMARY

Body summary

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

dependencies are tracked in front matter.

## REQUIREMENTS

requirements

## ACCEPTANCE

acceptance

## EDGE-CASES

not applicable

## OPEN-QUESTIONS

not required
`, "SPEC.md", TypeSpec)

	if !doc.FrontMatterPresent {
		t.Fatal("FrontMatterPresent = false, want true")
	}
	if got := doc.SummaryText(); got != "Metadata summary" {
		t.Fatalf("SummaryText() = %q, want metadata summary", got)
	}
	if section := doc.GetSection("SUMMARY"); section == nil || section.Content != "Body summary" {
		t.Fatalf("SUMMARY section = %#v, want body section", section)
	}
	if errors := doc.Validate(); len(errors) != 0 {
		t.Fatalf("Validate() errors = %#v, want none", errors)
	}
}

func TestParseFrontMatterPlaceholderDetectionUsesBodyOnly(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
summary: "<!-- TODO: metadata-only placeholder -->"
---
# SPEC

## SUMMARY

done
`, "SPEC.md", TypeSpec)

	if doc.HasUnresolvedPlaceholders() {
		t.Fatal("HasUnresolvedPlaceholders() = true, want false for front-matter-only TODO")
	}
}

func TestParseDelimiterInBodyIsNotFrontMatter(t *testing.T) {
	doc := Parse(`# NOTE

---
metadata-looking body text
---

## SUMMARY

summary
`, "NOTE.md", TypeAnalysis)

	if doc.FrontMatterPresent {
		t.Fatal("FrontMatterPresent = true, want false")
	}
	if doc.GetSection("SUMMARY") == nil {
		t.Fatal("SUMMARY section missing")
	}
}

func TestValidateRejectsMalformedFrontMatter(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: [
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

	var found bool
	for _, err := range doc.Validate() {
		if err.Section == "FRONT MATTER" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Validate() did not report front matter error: %#v", doc.Validate())
	}
}
