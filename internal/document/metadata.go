package document

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	MetadataVersion = 1

	ArtifactBrainstorm = "brainstorm"
	ArtifactSpec       = "spec"
	ArtifactPlan       = "plan"
	ArtifactTasks      = "tasks"

	RelationshipBuildsOn  = "builds_on"
	RelationshipDependsOn = "depends_on"
	RelationshipRelatedTo = "related_to"

	DependencyStatusActive   = "active"
	DependencyStatusOptional = "optional"
	DependencyStatusStale    = "stale"
)

type MetadataDiagnosticSeverity string

const (
	MetadataDiagnosticError   MetadataDiagnosticSeverity = "error"
	MetadataDiagnosticWarning MetadataDiagnosticSeverity = "warning"
)

type FeatureMetadata struct {
	ID   string `yaml:"id,omitempty"`
	Slug string `yaml:"slug,omitempty"`
	Dir  string `yaml:"dir,omitempty"`
}

type MetadataRelationship struct {
	Type   string `yaml:"type"`
	Target string `yaml:"target"`
}

type MetadataDependency struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Location string `yaml:"location"`
	UsedFor  string `yaml:"used_for"`
	Status   string `yaml:"status"`
}

type MetadataSkill struct {
	Name     string `yaml:"name"`
	Source   string `yaml:"source"`
	Path     string `yaml:"path"`
	Trigger  string `yaml:"trigger"`
	Required bool   `yaml:"required"`
}

type Metadata struct {
	KitMetadataVersion int                    `yaml:"kit_metadata_version"`
	Artifact           string                 `yaml:"artifact"`
	Feature            FeatureMetadata        `yaml:"feature"`
	Summary            string                 `yaml:"summary,omitempty"`
	Intent             string                 `yaml:"intent,omitempty"`
	Relationships      []MetadataRelationship `yaml:"relationships,omitempty"`
	Dependencies       []MetadataDependency   `yaml:"dependencies,omitempty"`
	Skills             []MetadataSkill        `yaml:"skills,omitempty"`
}

type MetadataDiagnostic struct {
	Severity MetadataDiagnosticSeverity
	Field    string
	Message  string
	Fix      string
}

type MetadataConflict struct {
	Field      string
	FrontValue string
	BodyValue  string
	Message    string
}

type metadataBlock struct {
	Raw           string
	Body          string
	Present       bool
	BodyStartLine int
	Err           error
}

type MetadataUpsert struct {
	Feature       FeatureMetadata
	Artifact      string
	Summary       string
	Intent        string
	Relationships []MetadataRelationship
	Dependencies  []MetadataDependency
	Skills        []MetadataSkill
}

var featureDirPattern = regexp.MustCompile(`^([0-9]+)-(.+)$`)

func FeatureMetadataFromDir(dirName string) FeatureMetadata {
	matches := featureDirPattern.FindStringSubmatch(strings.TrimSpace(dirName))
	if matches == nil {
		return FeatureMetadata{Dir: strings.TrimSpace(dirName)}
	}
	return FeatureMetadata{
		ID:   matches[1],
		Slug: matches[2],
		Dir:  strings.TrimSpace(dirName),
	}
}

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

	for i, dependency := range metadata.Dependencies {
		field := fmt.Sprintf("dependencies[%d]", i)
		if strings.TrimSpace(dependency.Name) == "" {
			diagnostics = append(diagnostics, metadataError(field+".name", "dependency name cannot be empty", "set dependency name"))
		}
		if strings.TrimSpace(dependency.Status) == "" || !isValidDependencyStatus(dependency.Status) {
			diagnostics = append(diagnostics, metadataError(
				field+".status",
				fmt.Sprintf("invalid dependency status %q", dependency.Status),
				"set dependency status to one of: active, optional, stale",
			))
		}
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

func isValidDependencyStatus(value string) bool {
	switch value {
	case DependencyStatusActive, DependencyStatusOptional, DependencyStatusStale:
		return true
	default:
		return false
	}
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
	conflicts = append(conflicts, d.dependencyConflicts()...)
	conflicts = append(conflicts, d.skillConflicts()...)
	return conflicts
}

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
	if len(update.Dependencies) > 0 {
		upsertDependencies(metadataNode, update.Dependencies)
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

func upsertDependencies(parent *yaml.Node, dependencies []MetadataDependency) {
	seq := findOrCreateSequence(parent, "dependencies")
	for _, dependency := range dependencies {
		existing := findDependencyNode(seq, dependency)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setNodeString(existing, "name", dependency.Name)
		setNodeString(existing, "type", dependency.Type)
		setNodeString(existing, "location", dependency.Location)
		setNodeString(existing, "used_for", dependency.UsedFor)
		setNodeString(existing, "status", dependency.Status)
	}
}

func findDependencyNode(seq *yaml.Node, dependency MetadataDependency) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if strings.EqualFold(getNodeString(item, "name"), dependency.Name) &&
			strings.EqualFold(getNodeString(item, "type"), dependency.Type) &&
			getNodeString(item, "location") == dependency.Location {
			return item
		}
	}
	return nil
}

func upsertSkills(parent *yaml.Node, skills []MetadataSkill) {
	seq := findOrCreateSequence(parent, "skills")
	for _, skill := range skills {
		existing := findSkillNode(seq, skill)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setNodeString(existing, "name", skill.Name)
		setNodeString(existing, "source", skill.Source)
		setNodeString(existing, "path", skill.Path)
		setNodeString(existing, "trigger", skill.Trigger)
		setNodeBool(existing, "required", skill.Required)
	}
}

func findSkillNode(seq *yaml.Node, skill MetadataSkill) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if strings.EqualFold(getNodeString(item, "name"), skill.Name) && getNodeString(item, "path") == skill.Path {
			return item
		}
	}
	return nil
}

func findOrCreateMapping(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		if parent.Content[i+1].Kind != yaml.MappingNode {
			parent.Content[i+1] = &yaml.Node{Kind: yaml.MappingNode}
		}
		return parent.Content[i+1]
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode
}

func findOrCreateSequence(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		if parent.Content[i+1].Kind != yaml.SequenceNode {
			parent.Content[i+1] = &yaml.Node{Kind: yaml.SequenceNode}
		}
		return parent.Content[i+1]
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.SequenceNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode
}

func setNodeString(parent *yaml.Node, key, value string) {
	setNode(parent, key, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
}

func setNodeInt(parent *yaml.Node, key string, value int) {
	setNode(parent, key, &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", value)})
}

func setNodeBool(parent *yaml.Node, key string, value bool) {
	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "false"}
	if value {
		scalar.Value = "true"
	}
	setNode(parent, key, scalar)
}

func setNode(parent *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = value
			return
		}
	}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
}

func getNodeString(parent *yaml.Node, key string) string {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1].Value
		}
	}
	return ""
}
