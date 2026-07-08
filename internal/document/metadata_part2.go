package document

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func ArtifactForDocumentType(docType DocumentType) string {
	switch docType {
	case TypeBrainstorm:
		return ArtifactBrainstorm
	case TypeSpec:
		return ArtifactSpec
	case TypePlan:
		return ArtifactPlan
	case TypeTasks:
		return ArtifactTasks
	default:
		return ""
	}
}

func splitLeadingFrontMatter(content string) metadataBlock {
	block := metadataBlock{
		Body:          content,
		BodyStartLine: 1,
	}
	if content == "" {
		return block
	}

	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return block
	}

	block.Present = true
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "---" {
			continue
		}
		block.Raw = strings.Join(lines[1:i], "")
		block.Body = strings.Join(lines[i+1:], "")
		block.BodyStartLine = i + 2
		return block
	}

	block.Raw = strings.Join(lines[1:], "")
	block.Body = content
	block.Err = fmt.Errorf("missing closing front matter delimiter")
	return block
}

func parseMetadata(raw string, docType DocumentType) (*Metadata, []MetadataDiagnostic) {
	var metadata Metadata
	var diagnostics []MetadataDiagnostic

	if strings.TrimSpace(raw) == "" {
		diagnostics = append(diagnostics, MetadataDiagnostic{
			Severity: MetadataDiagnosticError,
			Field:    "front_matter",
			Message:  "front matter is empty",
			Fix:      "populate the YAML front matter block or remove the delimiters",
		})
		return &metadata, diagnostics
	}

	if err := yaml.Unmarshal([]byte(raw), &metadata); err != nil {
		diagnostics = append(diagnostics, MetadataDiagnostic{
			Severity: MetadataDiagnosticError,
			Field:    "front_matter",
			Message:  fmt.Sprintf("failed to parse front matter: %v", err),
			Fix:      "fix the YAML front matter syntax before running validation again",
		})
		return &metadata, diagnostics
	}

	diagnostics = append(diagnostics, validateMetadata(metadata, docType)...)
	return &metadata, diagnostics
}

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
		if strings.TrimSpace(reference.Relation) == "" || !isValidReferenceRelation(reference.Relation) {
			diagnostics = append(diagnostics, metadataError(
				field+".relation",
				fmt.Sprintf("invalid reference relation %q", reference.Relation),
				"set reference relation to one of: constrains, supports, implements, verifies, guides, informs, supersedes, conflicts_with, uses",
			))
		}
		if strings.TrimSpace(reference.ReadPolicy) == "" || !isValidReferenceReadPolicy(reference.ReadPolicy) {
			diagnostics = append(diagnostics, metadataError(
				field+".read_policy",
				fmt.Sprintf("invalid reference read_policy %q", reference.ReadPolicy),
				"set reference read_policy to one of: must, conditional, evidence, skip",
			))
		}
		if strings.TrimSpace(reference.Status) == "" || !isValidReferenceStatus(reference.Status) {
			diagnostics = append(diagnostics, metadataError(
				field+".status",
				fmt.Sprintf("invalid reference status %q", reference.Status),
				"set reference status to one of: active, optional, stale",
			))
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
