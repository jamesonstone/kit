package document

import (
	"regexp"
	"strings"
)

const (
	MetadataVersion = 1

	ArtifactBrainstorm = "brainstorm"
	ArtifactSpec       = "spec"
	ArtifactPlan       = "plan"
	ArtifactTasks      = "tasks"

	ClarificationStatusOpen    = "open"
	ClarificationStatusReady   = "ready"
	ClarificationStatusBlocked = "blocked"

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

type MetadataClarification struct {
	Status              string `yaml:"status,omitempty"`
	Confidence          *int   `yaml:"confidence,omitempty"`
	UnresolvedQuestions *int   `yaml:"unresolved_questions,omitempty"`
}

type Metadata struct {
	KitMetadataVersion int                            `yaml:"kit_metadata_version"`
	Artifact           string                         `yaml:"artifact"`
	WorkflowVersion    int                            `yaml:"workflow_version,omitempty"`
	Phase              string                         `yaml:"phase,omitempty"`
	DeliveryIntent     string                         `yaml:"delivery_intent,omitempty"`
	Clarification      *MetadataClarification         `yaml:"clarification,omitempty"`
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
	DeliveryIntent  string
	Clarification   *MetadataClarification
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

func NewMetadataClarification(status string, confidence int, unresolvedQuestions int) MetadataClarification {
	return MetadataClarification{
		Status:              status,
		Confidence:          intPtr(confidence),
		UnresolvedQuestions: intPtr(unresolvedQuestions),
	}
}

func (c MetadataClarification) ConfidenceValue() (int, bool) {
	if c.Confidence == nil {
		return 0, false
	}
	return *c.Confidence, true
}

func (c MetadataClarification) UnresolvedQuestionsValue() (int, bool) {
	if c.UnresolvedQuestions == nil {
		return 0, false
	}
	return *c.UnresolvedQuestions, true
}

func intPtr(value int) *int {
	return &value
}
