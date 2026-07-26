package document

import (
	"strings"
	"testing"
)

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
