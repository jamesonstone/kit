package document

import (
	"fmt"
	"regexp"
	"strings"
)

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

func isValidReferenceReadPolicy(value string) bool {
	switch value {
	case ReferenceReadPolicyMust,
		ReferenceReadPolicyConditional,
		ReferenceReadPolicyEvidence,
		ReferenceReadPolicySkip:
		return true
	default:
		return false
	}
}

func isValidReferenceSelectorType(value string) bool {
	switch value {
	case ReferenceSelectorTypeArtifact,
		ReferenceSelectorTypeHeading,
		ReferenceSelectorTypeSymbol,
		ReferenceSelectorTypeCommand,
		ReferenceSelectorTypeURL,
		ReferenceSelectorTypeNodeID:
		return true
	default:
		return false
	}
}

func referencePolicyDiagnostics(field string, reference MetadataReference) []MetadataDiagnostic {
	var diagnostics []MetadataDiagnostic
	relation := strings.TrimSpace(reference.Relation)
	readPolicy := strings.TrimSpace(reference.ReadPolicy)
	status := strings.TrimSpace(reference.Status)

	if status == ReferenceStatusStale && readPolicy != ReferenceReadPolicySkip {
		diagnostics = append(diagnostics, metadataWarning(
			field+".read_policy",
			"stale reference should normally be skipped",
			"set read_policy: skip or change status if the reference is still active",
		))
	}
	if status == ReferenceStatusActive && readPolicy == ReferenceReadPolicySkip {
		diagnostics = append(diagnostics, metadataWarning(
			field+".status",
			"active reference is marked skip",
			"set status: stale or optional unless the active reference should be excluded from context plans",
		))
	}
	if relation == ReferenceRelationConstrains && readPolicy != ReferenceReadPolicyMust {
		diagnostics = append(diagnostics, metadataWarning(
			field+".read_policy",
			"constraining reference should normally be must-read",
			"set read_policy: must unless the constraint is only conditionally relevant",
		))
	}
	if relation == ReferenceRelationVerifies && readPolicy != ReferenceReadPolicyEvidence {
		diagnostics = append(diagnostics, metadataWarning(
			field+".read_policy",
			"verification reference should normally be evidence-read",
			"set read_policy: evidence unless the reference is needed before implementation",
		))
	}

	return diagnostics
}

func hasUnpinnedLineReference(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if strings.Contains(value, "/blob/") && regexp.MustCompile(`/blob/[0-9a-f]{7,40}/`).MatchString(value) {
		return false
	}
	return regexp.MustCompile(`(^|[./_-])[A-Za-z0-9_./-]+\.[A-Za-z0-9]+:(L)?[0-9]+(-L?[0-9]+)?$`).MatchString(value)
}

func RelationshipHumanToMachine(value string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "builds on", RelationshipBuildsOn:
		return RelationshipBuildsOn, true
	case "depends on", RelationshipDependsOn:
		return RelationshipDependsOn, true
	case "related to", RelationshipRelatedTo:
		return RelationshipRelatedTo, true
	default:
		return "", false
	}
}

func RelationshipMachineToHuman(value string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case RelationshipBuildsOn, "builds on":
		return "builds on", true
	case RelationshipDependsOn, "depends on":
		return "depends on", true
	case RelationshipRelatedTo, "related to":
		return "related to", true
	default:
		return "", false
	}
}
