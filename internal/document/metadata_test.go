package document

import (
	"strings"
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
dependencies:
  - name: Thing
    type: doc
    location: docs/thing.md
    used_for: context
    status: maybe
---
# SPEC

## SUMMARY

summary
`, "SPEC.md", TypeSpec)

	var relationshipError, dependencyError bool
	for _, err := range doc.Validate() {
		if strings.Contains(err.Message, "invalid relationship type") {
			relationshipError = true
		}
		if strings.Contains(err.Message, "invalid dependency status") {
			dependencyError = true
		}
	}
	if !relationshipError || !dependencyError {
		t.Fatalf("Validate() relationship error = %v, dependency error = %v, errors = %#v", relationshipError, dependencyError, doc.Validate())
	}
}

func TestMetadataAccessorsPreferFrontMatterAndReportConflicts(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
relationships:
  - type: depends_on
    target: 0002-beta
dependencies:
  - name: Front
    type: doc
    location: docs/front.md
    used_for: front matter
    status: active
skills:
  - name: rlm
    source: repo-local doc
    path: docs/agents/RLM.md
    trigger: broad context
    required: true
---
# SPEC

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| legacy | doc | docs/legacy.md | legacy | no |

## RELATIONSHIPS

- depends on: 0003-gamma

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Legacy | doc | docs/legacy.md | legacy | active |
`, "SPEC.md", TypeSpec)

	relationships, warnings := doc.Relationships()
	if len(warnings) != 0 {
		t.Fatalf("Relationships() warnings = %#v, want none", warnings)
	}
	if len(relationships) != 1 || relationships[0].Type != "depends on" || relationships[0].Target != "0002-beta" {
		t.Fatalf("Relationships() = %#v, want front matter relationship", relationships)
	}
	if got := doc.Dependencies(); len(got) != 1 || got[0].Name != "Front" {
		t.Fatalf("Dependencies() = %#v, want front matter dependency", got)
	}
	if got := doc.Skills(); len(got) != 1 || got[0].Name != "rlm" || !got[0].Required {
		t.Fatalf("Skills() = %#v, want front matter skill", got)
	}
	if len(doc.MetadataConflictWarnings) != 3 {
		t.Fatalf("MetadataConflictWarnings len = %d, want 3: %#v", len(doc.MetadataConflictWarnings), doc.MetadataConflictWarnings)
	}
}

func TestMetadataAccessorsFallbackToLegacySections(t *testing.T) {
	doc := Parse(`# SPEC

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| rlm | repo-local doc | docs/agents/RLM.md | broad context | yes |

## RELATIONSHIPS

- depends on: `+"`0002-beta`"+`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Legacy | doc | docs/legacy.md | legacy | active |
`, "SPEC.md", TypeSpec)

	relationships, warnings := doc.Relationships()
	if len(warnings) != 0 {
		t.Fatalf("Relationships() warnings = %#v, want none", warnings)
	}
	if len(relationships) != 1 || relationships[0].Target != "0002-beta" {
		t.Fatalf("Relationships() = %#v, want legacy relationship", relationships)
	}
	if got := doc.Dependencies(); len(got) != 1 || got[0].Name != "Legacy" {
		t.Fatalf("Dependencies() = %#v, want legacy dependency", got)
	}
	if got := doc.Skills(); len(got) != 1 || got[0].Name != "rlm" || !got[0].Required {
		t.Fatalf("Skills() = %#v, want legacy skill", got)
	}
}

func TestUpsertMetadataPreservesUnknownFieldsAndBody(t *testing.T) {
	content := `---
custom_field: keep me
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
---
# SPEC

## SUMMARY

summary
`

	updated, changed, err := UpsertMetadata(content, TypeSpec, MetadataUpsert{
		Dependencies: []MetadataDependency{{
			Name:     "Feature notes",
			Type:     "notes",
			Location: "docs/notes/0001-alpha",
			UsedFor:  "optional pre-brainstorm input",
			Status:   DependencyStatusOptional,
		}},
	})
	if err != nil {
		t.Fatalf("UpsertMetadata() error = %v", err)
	}
	if !changed {
		t.Fatal("UpsertMetadata() changed = false, want true")
	}
	if !strings.Contains(updated, "custom_field: keep me") {
		t.Fatalf("updated content lost unknown field:\n%s", updated)
	}
	if !strings.Contains(updated, "# SPEC\n\n## SUMMARY\n\nsummary\n") {
		t.Fatalf("updated content lost body:\n%s", updated)
	}

	doc := Parse(updated, "SPEC.md", TypeSpec)
	if got := doc.Dependencies(); len(got) != 1 || got[0].Name != "Feature notes" {
		t.Fatalf("Dependencies() = %#v, want upserted dependency", got)
	}
}

func TestUpsertMetadataRejectsUnclosedFrontMatter(t *testing.T) {
	content := `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
# SPEC
`

	_, changed, err := UpsertMetadata(content, TypeSpec, MetadataUpsert{
		Dependencies: []MetadataDependency{{
			Name:     "Feature notes",
			Type:     "notes",
			Location: "docs/notes/0001-alpha",
			UsedFor:  "optional pre-brainstorm input",
			Status:   DependencyStatusOptional,
		}},
	})
	if err == nil {
		t.Fatal("UpsertMetadata() error = nil, want unclosed front matter error")
	}
	if changed {
		t.Fatal("UpsertMetadata() changed = true, want false on parse error")
	}
}
