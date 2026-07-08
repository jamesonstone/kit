package document

import (
	"strings"
	"testing"
)

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
references:
  - name: Front
    type: doc
    target: docs/front.md
    relation: informs
    read_policy: conditional
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
	if got := doc.References(); len(got) != 1 || got[0].Name != "Front" || got[0].Target != "docs/front.md" {
		t.Fatalf("References() = %#v, want front matter reference", got)
	}
	if got := doc.Skills(); len(got) != 1 || got[0].Name != "rlm" || !got[0].Required {
		t.Fatalf("Skills() = %#v, want front matter skill", got)
	}
	if len(doc.MetadataConflictWarnings) != 2 {
		t.Fatalf("MetadataConflictWarnings len = %d, want 2: %#v", len(doc.MetadataConflictWarnings), doc.MetadataConflictWarnings)
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
	if got := doc.References(); len(got) != 0 {
		t.Fatalf("References() = %#v, want no legacy dependency fallback", got)
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
		DeliveryIntent: "idea_only",
		References: []MetadataReference{{
			ID:         "feature-notes",
			Name:       "Feature notes",
			Type:       "notes",
			Target:     "docs/notes/0001-alpha",
			Relation:   ReferenceRelationInforms,
			ReadPolicy: ReferenceReadPolicyConditional,
			UsedFor:    "optional pre-brainstorm input",
			Status:     ReferenceStatusOptional,
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
	if !strings.Contains(updated, "delivery_intent: idea_only") {
		t.Fatalf("updated content missing delivery intent:\n%s", updated)
	}

	doc := Parse(updated, "SPEC.md", TypeSpec)
	if got := doc.DeliveryIntent(); got != "idea_only" {
		t.Fatalf("DeliveryIntent() = %q, want idea_only", got)
	}
	if got := doc.References(); len(got) != 1 || got[0].Name != "Feature notes" || got[0].Target != "docs/notes/0001-alpha" {
		t.Fatalf("References() = %#v, want upserted reference", got)
	}
	if got := doc.References()[0].ID; got != "feature-notes" {
		t.Fatalf("reference ID = %q, want feature-notes", got)
	}
}

func TestUpsertMetadataAddsClarificationState(t *testing.T) {
	content := `---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: clarify
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
---
# SPEC

## THESIS

alpha
`
	clarification := NewMetadataClarification(ClarificationStatusOpen, 0, 1)
	updated, changed, err := UpsertMetadata(content, TypeSpec, MetadataUpsert{
		Clarification: &clarification,
	})
	if err != nil {
		t.Fatalf("UpsertMetadata() error = %v", err)
	}
	if !changed {
		t.Fatal("UpsertMetadata() changed = false, want true")
	}
	for _, check := range []string{
		"clarification:",
		"  status: open",
		"  confidence: 0",
		"  unresolved_questions: 1",
	} {
		if !strings.Contains(updated, check) {
			t.Fatalf("expected updated metadata to contain %q, got:\n%s", check, updated)
		}
	}
	doc := Parse(updated, "SPEC.md", TypeSpec)
	got, ok := doc.ClarificationState()
	if !ok {
		t.Fatalf("ClarificationState() ok = false")
	}
	if got.Status != ClarificationStatusOpen {
		t.Fatalf("clarification status = %q, want open", got.Status)
	}
	confidence, ok := got.ConfidenceValue()
	if !ok || confidence != 0 {
		t.Fatalf("clarification confidence = %d, %v; want 0, true", confidence, ok)
	}
	unresolved, ok := got.UnresolvedQuestionsValue()
	if !ok || unresolved != 1 {
		t.Fatalf("clarification unresolved = %d, %v; want 1, true", unresolved, ok)
	}
}
