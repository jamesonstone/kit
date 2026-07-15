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

func TestMetadataWarnsWhenClarificationStateMissingFromV2Spec(t *testing.T) {
	doc := Parse(`---
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
`, "SPEC.md", TypeSpec)

	var found bool
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity == MetadataDiagnosticWarning &&
			diagnostic.Field == "clarification" &&
			strings.Contains(diagnostic.Message, "clarification state") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing clarification warning, got %#v", doc.MetadataDiagnostics)
	}
}

func TestMetadataRejectsInvalidClarificationState(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: clarify
clarification:
  status: done
  confidence: 101
  unresolved_questions: -1
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
---
# SPEC
`, "SPEC.md", TypeSpec)

	var invalidStatus, invalidConfidence, invalidUnresolved bool
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity != MetadataDiagnosticError {
			continue
		}
		switch diagnostic.Field {
		case "clarification.status":
			invalidStatus = true
		case "clarification.confidence":
			invalidConfidence = true
		case "clarification.unresolved_questions":
			invalidUnresolved = true
		}
	}
	if !invalidStatus || !invalidConfidence || !invalidUnresolved {
		t.Fatalf("expected invalid clarification errors, got %#v", doc.MetadataDiagnostics)
	}
}

func TestMetadataAcceptsLegacyV2ReferenceVocabularyWithWarnings(t *testing.T) {
	doc := Parse(`---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: deliver
clarification:
  status: ready
  confidence: 100
  unresolved_questions: 0
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
references:
  - name: Legacy guidance
    type: doc
    target: docs/agents/README.md
    relation: governs
    read_policy: must
    used_for: compatibility
    status: loaded
---
# SPEC
`, "SPEC.md", TypeSpec)

	var governsWarning, loadedWarning bool
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity == MetadataDiagnosticError {
			t.Fatalf("unexpected legacy compatibility error: %#v", diagnostic)
		}
		governsWarning = governsWarning || strings.Contains(diagnostic.Message, `relation "governs"`)
		loadedWarning = loadedWarning || strings.Contains(diagnostic.Message, `status "loaded"`)
	}
	if !governsWarning || !loadedWarning {
		t.Fatalf("legacy warnings governs=%v loaded=%v diagnostics=%#v", governsWarning, loadedWarning, doc.MetadataDiagnostics)
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
		References: []MetadataReference{{
			Name:       "Feature notes",
			Type:       "notes",
			Target:     "docs/notes/0001-alpha",
			Relation:   ReferenceRelationInforms,
			ReadPolicy: ReferenceReadPolicyConditional,
			UsedFor:    "optional pre-brainstorm input",
			Status:     ReferenceStatusOptional,
		}},
	})
	if err == nil {
		t.Fatal("UpsertMetadata() error = nil, want unclosed front matter error")
	}
	if changed {
		t.Fatal("UpsertMetadata() changed = true, want false on parse error")
	}
}
