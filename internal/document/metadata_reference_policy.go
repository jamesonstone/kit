package document

import (
	"regexp"
	"strings"
)

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

func (d *Document) HasFrontMatterErrors() bool {
	for _, diagnostic := range d.MetadataDiagnostics {
		if diagnostic.Severity == MetadataDiagnosticError {
			return true
		}
	}
	return false
}

func (d *Document) metadataConflicts() []MetadataConflict {
	if d.Metadata == nil {
		return nil
	}

	var conflicts []MetadataConflict
	conflicts = append(conflicts, d.relationshipConflicts()...)
	conflicts = append(conflicts, d.skillConflicts()...)
	return conflicts
}
