package document

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func UpsertMetadata(content string, docType DocumentType, update MetadataUpsert) (string, bool, error) {
	block := splitLeadingFrontMatter(content)
	if block.Err != nil {
		return content, false, fmt.Errorf("failed to parse front matter for update: %w", block.Err)
	}
	metadataNode, err := metadataMappingNode(block.Raw)
	if err != nil {
		return content, false, err
	}

	if update.Artifact == "" {
		update.Artifact = ArtifactForDocumentType(docType)
	}

	setNodeInt(metadataNode, "kit_metadata_version", MetadataVersion)
	if update.Artifact != "" {
		setNodeString(metadataNode, "artifact", update.Artifact)
	}
	if update.WorkflowVersion != 0 {
		setNodeInt(metadataNode, "workflow_version", update.WorkflowVersion)
	}
	if update.Phase != "" {
		setNodeString(metadataNode, "phase", update.Phase)
	}
	if update.DeliveryIntent != "" {
		setNodeString(metadataNode, "delivery_intent", update.DeliveryIntent)
	}
	if update.Clarification != nil {
		upsertClarification(metadataNode, *update.Clarification)
	}
	if update.Feature != (FeatureMetadata{}) {
		upsertFeatureMetadata(metadataNode, update.Feature)
	}
	if update.Summary != "" {
		setNodeString(metadataNode, "summary", update.Summary)
	}
	if update.Intent != "" {
		setNodeString(metadataNode, "intent", update.Intent)
	}
	if len(update.Relationships) > 0 {
		upsertRelationships(metadataNode, update.Relationships)
	}
	if len(update.References) > 0 {
		upsertReferences(metadataNode, update.References)
	}
	if len(update.Skills) > 0 {
		upsertSkills(metadataNode, update.Skills)
	}

	encoded, err := encodeMetadataNode(metadataNode)
	if err != nil {
		return content, false, err
	}

	body := block.Body
	if !block.Present {
		body = content
	}
	updated := "---\n" + encoded + "---\n" + body
	if updated == content {
		return content, false, nil
	}
	return updated, true, nil
}

func metadataMappingNode(raw string) (*yaml.Node, error) {
	if strings.TrimSpace(raw) == "" {
		return &yaml.Node{Kind: yaml.MappingNode}, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse front matter for update: %w", err)
	}
	if len(doc.Content) == 0 {
		return &yaml.Node{Kind: yaml.MappingNode}, nil
	}
	if doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("front matter root must be a YAML mapping")
	}
	return doc.Content[0], nil
}

func encodeMetadataNode(node *yaml.Node) (string, error) {
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		_ = encoder.Close()
		return "", fmt.Errorf("failed to encode front matter: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to encode front matter: %w", err)
	}
	return output.String(), nil
}

func upsertFeatureMetadata(parent *yaml.Node, feature FeatureMetadata) {
	featureNode := findOrCreateMapping(parent, "feature")
	if feature.ID != "" {
		setNodeString(featureNode, "id", feature.ID)
	}
	if feature.Slug != "" {
		setNodeString(featureNode, "slug", feature.Slug)
	}
	if feature.Dir != "" {
		setNodeString(featureNode, "dir", feature.Dir)
	}
}

func upsertClarification(parent *yaml.Node, clarification MetadataClarification) {
	clarificationNode := findOrCreateMapping(parent, "clarification")
	if clarification.Status != "" {
		setNodeString(clarificationNode, "status", clarification.Status)
	}
	if clarification.Confidence != nil {
		setNodeInt(clarificationNode, "confidence", *clarification.Confidence)
	}
	if clarification.UnresolvedQuestions != nil {
		setNodeInt(clarificationNode, "unresolved_questions", *clarification.UnresolvedQuestions)
	}
}

func upsertRelationships(parent *yaml.Node, relationships []MetadataRelationship) {
	seq := findOrCreateSequence(parent, "relationships")
	for _, relationship := range relationships {
		existing := findRelationshipNode(seq, relationship)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setNodeString(existing, "type", relationship.Type)
		setNodeString(existing, "target", relationship.Target)
	}
}

func findRelationshipNode(seq *yaml.Node, relationship MetadataRelationship) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if getNodeString(item, "type") == relationship.Type && getNodeString(item, "target") == relationship.Target {
			return item
		}
	}
	return nil
}
