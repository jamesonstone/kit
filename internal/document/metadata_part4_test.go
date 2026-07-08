package document

import (
	"strings"
	"testing"
)

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
