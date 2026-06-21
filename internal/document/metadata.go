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

	ReferenceRelationConstrains    = "constrains"
	ReferenceRelationSupports      = "supports"
	ReferenceRelationImplements    = "implements"
	ReferenceRelationVerifies      = "verifies"
	ReferenceRelationGuides        = "guides"
	ReferenceRelationInforms       = "informs"
	ReferenceRelationSupersedes    = "supersedes"
	ReferenceRelationConflictsWith = "conflicts_with"
	ReferenceRelationUses          = "uses"

	ReferenceReadPolicyMust        = "must"
	ReferenceReadPolicyConditional = "conditional"
	ReferenceReadPolicyEvidence    = "evidence"
	ReferenceReadPolicySkip        = "skip"

	ReferenceStatusActive   = "active"
	ReferenceStatusOptional = "optional"
	ReferenceStatusStale    = "stale"

	ReferenceSelectorTypeArtifact = "artifact"
	ReferenceSelectorTypeHeading  = "heading"
	ReferenceSelectorTypeSymbol   = "symbol"
	ReferenceSelectorTypeCommand  = "command"
	ReferenceSelectorTypeURL      = "url"
	ReferenceSelectorTypeNodeID   = "node_id"
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

type deprecatedMetadataDependency struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Location string `yaml:"location"`
	UsedFor  string `yaml:"used_for"`
	Status   string `yaml:"status"`
}

type MetadataReference struct {
	ID           string `yaml:"id,omitempty"`
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	Target       string `yaml:"target"`
	SelectorType string `yaml:"selector_type,omitempty"`
	Selector     string `yaml:"selector,omitempty"`
	Relation     string `yaml:"relation"`
	ReadPolicy   string `yaml:"read_policy"`
	UsedFor      string `yaml:"used_for"`
	Status       string `yaml:"status"`
}

type MetadataSkill struct {
	Name     string `yaml:"name"`
	Source   string `yaml:"source"`
	Path     string `yaml:"path"`
	Trigger  string `yaml:"trigger"`
	Required bool   `yaml:"required"`
}

type Metadata struct {
	KitMetadataVersion int                            `yaml:"kit_metadata_version"`
	Artifact           string                         `yaml:"artifact"`
	WorkflowVersion    int                            `yaml:"workflow_version,omitempty"`
	Phase              string                         `yaml:"phase,omitempty"`
	Feature            FeatureMetadata                `yaml:"feature"`
	Summary            string                         `yaml:"summary,omitempty"`
	Intent             string                         `yaml:"intent,omitempty"`
	Relationships      []MetadataRelationship         `yaml:"relationships,omitempty"`
	References         []MetadataReference            `yaml:"references,omitempty"`
	Dependencies       []deprecatedMetadataDependency `yaml:"dependencies,omitempty"`
	Skills             []MetadataSkill                `yaml:"skills,omitempty"`
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
	Feature         FeatureMetadata
	Artifact        string
	WorkflowVersion int
	Phase           string
	Summary         string
	Intent          string
	Relationships   []MetadataRelationship
	References      []MetadataReference
	Skills          []MetadataSkill
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

	if len(metadata.Dependencies) > 0 {
		diagnostics = append(diagnostics, metadataError(
			"dependencies",
			"front matter dependencies are deprecated",
			"migrate `dependencies` entries to canonical `references` entries with `target`, `relation`, and `read_policy`",
		))
	}

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

func upsertReferences(parent *yaml.Node, references []MetadataReference) {
	seq := findOrCreateSequence(parent, "references")
	for _, reference := range references {
		existing := findReferenceNode(seq, reference)
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			seq.Content = append(seq.Content, existing)
		}
		setOptionalNodeString(existing, "id", reference.ID)
		setNodeString(existing, "name", reference.Name)
		setNodeString(existing, "type", reference.Type)
		setNodeString(existing, "target", reference.Target)
		setOptionalNodeString(existing, "selector_type", reference.SelectorType)
		setOptionalNodeString(existing, "selector", reference.Selector)
		setNodeString(existing, "relation", reference.Relation)
		setNodeString(existing, "read_policy", reference.ReadPolicy)
		setNodeString(existing, "used_for", reference.UsedFor)
		setNodeString(existing, "status", reference.Status)
	}
}

func findReferenceNode(seq *yaml.Node, reference MetadataReference) *yaml.Node {
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		if strings.TrimSpace(reference.ID) != "" &&
			strings.EqualFold(getNodeString(item, "id"), reference.ID) {
			return item
		}
		if strings.EqualFold(getNodeString(item, "name"), reference.Name) &&
			strings.EqualFold(getNodeString(item, "type"), reference.Type) &&
			getNodeString(item, "target") == reference.Target &&
			getNodeString(item, "selector_type") == reference.SelectorType &&
			getNodeString(item, "selector") == reference.Selector {
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

func setOptionalNodeString(parent *yaml.Node, key, value string) {
	if strings.TrimSpace(value) == "" {
		removeNode(parent, key)
		return
	}
	setNodeString(parent, key, value)
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

func removeNode(parent *yaml.Node, key string) {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
		return
	}
}

func getNodeString(parent *yaml.Node, key string) string {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1].Value
		}
	}
	return ""
}
