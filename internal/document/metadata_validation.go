package document

import (
	"fmt"
	"strings"
)

func validateMetadata(metadata Metadata, docType DocumentType) []MetadataDiagnostic {
	if !isFeatureArtifactType(docType) {
		return nil
	}

	var diagnostics []MetadataDiagnostic
	if metadata.KitMetadataVersion != MetadataVersion {
		diagnostics = append(diagnostics, metadataError(
			"kit_metadata_version",
			fmt.Sprintf("front matter must set kit_metadata_version: %d", MetadataVersion),
			fmt.Sprintf("set `kit_metadata_version: %d`", MetadataVersion),
		))
	}

	expectedArtifact := ArtifactForDocumentType(docType)
	if metadata.Artifact == "" {
		diagnostics = append(diagnostics, metadataError(
			"artifact",
			"front matter must set artifact",
			fmt.Sprintf("set `artifact: %s`", expectedArtifact),
		))
	} else if !isValidArtifact(metadata.Artifact) {
		diagnostics = append(diagnostics, metadataError(
			"artifact",
			fmt.Sprintf("invalid artifact %q", metadata.Artifact),
			"set artifact to one of: brainstorm, spec, plan, tasks",
		))
	} else if expectedArtifact != "" && metadata.Artifact != expectedArtifact {
		diagnostics = append(diagnostics, metadataError(
			"artifact",
			fmt.Sprintf("front matter artifact %q does not match document type %q", metadata.Artifact, expectedArtifact),
			fmt.Sprintf("set `artifact: %s`", expectedArtifact),
		))
	}

	if metadata.Feature.ID == "" {
		diagnostics = append(diagnostics, metadataError("feature.id", "front matter must set feature.id", "set `feature.id` to the numeric feature id"))
	}
	if metadata.Feature.Slug == "" {
		diagnostics = append(diagnostics, metadataError("feature.slug", "front matter must set feature.slug", "set `feature.slug` to the feature slug"))
	}
	if metadata.Feature.Dir == "" {
		diagnostics = append(diagnostics, metadataError("feature.dir", "front matter must set feature.dir", "set `feature.dir` to the canonical feature directory"))
	}

	for i, relationship := range metadata.Relationships {
		field := fmt.Sprintf("relationships[%d]", i)
		if _, ok := RelationshipMachineToHuman(relationship.Type); !ok {
			diagnostics = append(diagnostics, metadataError(
				field+".type",
				fmt.Sprintf("invalid relationship type %q", relationship.Type),
				"set relationship type to one of: builds_on, depends_on, related_to",
			))
		}
		if strings.TrimSpace(relationship.Target) == "" {
			diagnostics = append(diagnostics, metadataError(field+".target", "relationship target cannot be empty", "set relationship target to a feature directory id"))
		}
	}

	if len(metadata.Dependencies) > 0 {
		diagnostics = append(diagnostics, metadataError(
			"dependencies",
			"front matter dependencies are deprecated",
			"migrate `dependencies` entries to canonical `references` entries with `target`, `relation`, and `read_policy`",
		))
	}

	diagnostics = append(diagnostics, validateClarificationMetadata(metadata, docType)...)

	for i, reference := range metadata.References {
		field := fmt.Sprintf("references[%d]", i)
		if strings.TrimSpace(reference.Name) == "" {
			diagnostics = append(diagnostics, metadataError(field+".name", "reference name cannot be empty", "set reference name"))
		}
		if strings.TrimSpace(reference.Type) == "" {
			diagnostics = append(diagnostics, metadataError(field+".type", "reference type cannot be empty", "set reference type"))
		}
		if strings.TrimSpace(reference.Target) == "" {
			diagnostics = append(diagnostics, metadataError(field+".target", "reference target cannot be empty", "set reference target"))
		}
		relation := strings.TrimSpace(reference.Relation)
		if relation == "" || !isValidReferenceRelation(relation) {
			if metadata.WorkflowVersion == WorkflowVersionV2 && relation == "governs" {
				diagnostics = append(diagnostics, metadataWarning(
					field+".relation",
					"legacy V2 reference relation \"governs\" is supported for compatibility",
					"semantically curate the reference to `constrains` or `guides` when the V2 spec is next reviewed",
				))
			} else {
				diagnostics = append(diagnostics, metadataError(
					field+".relation",
					fmt.Sprintf("invalid reference relation %q", reference.Relation),
					"set reference relation to one of: constrains, supports, implements, verifies, guides, informs, supersedes, conflicts_with, uses",
				))
			}
		}
		if strings.TrimSpace(reference.ReadPolicy) == "" || !isValidReferenceReadPolicy(reference.ReadPolicy) {
			diagnostics = append(diagnostics, metadataError(
				field+".read_policy",
				fmt.Sprintf("invalid reference read_policy %q", reference.ReadPolicy),
				"set reference read_policy to one of: must, conditional, evidence, skip",
			))
		}
		status := strings.TrimSpace(reference.Status)
		if status == "" || !isValidReferenceStatus(status) {
			if metadata.WorkflowVersion == WorkflowVersionV2 && status == "loaded" {
				diagnostics = append(diagnostics, metadataWarning(
					field+".status",
					"legacy V2 reference status \"loaded\" is supported for compatibility",
					"semantically curate the status to `active`, `optional`, or `stale` when the V2 spec is next reviewed",
				))
			} else {
				diagnostics = append(diagnostics, metadataError(
					field+".status",
					fmt.Sprintf("invalid reference status %q", reference.Status),
					"set reference status to one of: active, optional, stale",
				))
			}
		}
		if strings.TrimSpace(reference.SelectorType) != "" && !isValidReferenceSelectorType(reference.SelectorType) {
			diagnostics = append(diagnostics, metadataError(
				field+".selector_type",
				fmt.Sprintf("invalid reference selector_type %q", reference.SelectorType),
				"set reference selector_type to one of: artifact, heading, symbol, command, url, node_id",
			))
		}
		if strings.TrimSpace(reference.SelectorType) != "" && strings.TrimSpace(reference.Selector) == "" {
			diagnostics = append(diagnostics, metadataWarning(
				field+".selector",
				"reference selector_type is set without selector",
				"set selector or remove selector_type",
			))
		}
		if strings.TrimSpace(reference.Selector) != "" && strings.TrimSpace(reference.SelectorType) == "" {
			diagnostics = append(diagnostics, metadataWarning(
				field+".selector_type",
				"reference selector is set without selector_type",
				"set selector_type so tooling can resolve the selector deterministically",
			))
		}
		diagnostics = append(diagnostics, referencePolicyDiagnostics(field, reference)...)
		if hasUnpinnedLineReference(reference.Target) || hasUnpinnedLineReference(reference.Selector) {
			diagnostics = append(diagnostics, metadataWarning(
				field+".target",
				"reference appears to use an unpinned line number",
				"prefer a stable selector such as artifact id, heading, symbol, URL/node id, or a commit-pinned permalink",
			))
		}
	}

	return diagnostics
}

func validateClarificationMetadata(metadata Metadata, docType DocumentType) []MetadataDiagnostic {
	if docType != TypeSpec || metadata.WorkflowVersion != 2 {
		return nil
	}
	if metadata.Clarification == nil {
		return []MetadataDiagnostic{metadataWarning(
			"clarification",
			"v2 SPEC.md front matter should include clarification state",
			"run `kit spec <feature>` to backfill clarification.status, clarification.confidence, and clarification.unresolved_questions",
		)}
	}

	clarification := metadata.Clarification
	var diagnostics []MetadataDiagnostic
	switch strings.TrimSpace(clarification.Status) {
	case ClarificationStatusOpen, ClarificationStatusReady, ClarificationStatusBlocked:
	default:
		diagnostics = append(diagnostics, metadataError(
			"clarification.status",
			fmt.Sprintf("invalid clarification status %q", clarification.Status),
			"set clarification.status to one of: open, ready, blocked",
		))
	}
	if clarification.Confidence == nil {
		diagnostics = append(diagnostics, metadataWarning(
			"clarification.confidence",
			"v2 SPEC.md front matter should include clarification confidence",
			"set clarification.confidence to an integer from 0 to 100",
		))
	} else if *clarification.Confidence < 0 || *clarification.Confidence > 100 {
		diagnostics = append(diagnostics, metadataError(
			"clarification.confidence",
			fmt.Sprintf("clarification confidence %d is outside 0..100", *clarification.Confidence),
			"set clarification.confidence to an integer from 0 to 100",
		))
	}
	if clarification.UnresolvedQuestions == nil {
		diagnostics = append(diagnostics, metadataWarning(
			"clarification.unresolved_questions",
			"v2 SPEC.md front matter should include unresolved question count",
			"set clarification.unresolved_questions to an integer greater than or equal to 0",
		))
	} else if *clarification.UnresolvedQuestions < 0 {
		diagnostics = append(diagnostics, metadataError(
			"clarification.unresolved_questions",
			fmt.Sprintf("clarification unresolved question count %d is negative", *clarification.UnresolvedQuestions),
			"set clarification.unresolved_questions to an integer greater than or equal to 0",
		))
	}
	return diagnostics
}

func metadataError(field, message, fix string) MetadataDiagnostic {
	return MetadataDiagnostic{
		Severity: MetadataDiagnosticError,
		Field:    field,
		Message:  message,
		Fix:      fix,
	}
}

func metadataWarning(field, message, fix string) MetadataDiagnostic {
	return MetadataDiagnostic{
		Severity: MetadataDiagnosticWarning,
		Field:    field,
		Message:  message,
		Fix:      fix,
	}
}

func isFeatureArtifactType(docType DocumentType) bool {
	switch docType {
	case TypeBrainstorm, TypeSpec, TypePlan, TypeTasks:
		return true
	default:
		return false
	}
}

func isValidArtifact(value string) bool {
	switch value {
	case ArtifactBrainstorm, ArtifactSpec, ArtifactPlan, ArtifactTasks:
		return true
	default:
		return false
	}
}

func isValidReferenceStatus(value string) bool {
	switch value {
	case ReferenceStatusActive, ReferenceStatusOptional, ReferenceStatusStale:
		return true
	default:
		return false
	}
}

func isValidReferenceRelation(value string) bool {
	switch value {
	case ReferenceRelationConstrains,
		ReferenceRelationSupports,
		ReferenceRelationImplements,
		ReferenceRelationVerifies,
		ReferenceRelationGuides,
		ReferenceRelationInforms,
		ReferenceRelationSupersedes,
		ReferenceRelationConflictsWith,
		ReferenceRelationUses:
		return true
	default:
		return false
	}
}
