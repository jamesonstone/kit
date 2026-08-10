package contract

import "github.com/jamesonstone/kit/internal/registry"

type Hints struct {
	WorkType      string   `json:"work_type,omitempty"`
	Feature       string   `json:"feature,omitempty"`
	Paths         []string `json:"paths,omitempty"`
	Applicability []string `json:"applicability,omitempty"`
	Workflows     []string `json:"workflows,omitempty"`
}

type Resolved struct {
	SchemaVersion int                   `json:"schema_version"`
	State         string                `json:"state"`
	ProjectRoot   string                `json:"project_root"`
	Hints         Hints                 `json:"hints"`
	FeatureSpec   *FeatureSpec          `json:"feature_spec,omitempty"`
	Registry      registry.SourceConfig `json:"registry"`
	Routing       []string              `json:"routing"`
	Workflows     []Artifact            `json:"workflows"`
	Rules         map[string][]Artifact `json:"rules"`
	Diagnostics   []string              `json:"diagnostics,omitempty"`
	NextActions   []string              `json:"next_actions,omitempty"`
}

type FeatureSpec struct {
	Feature          string           `json:"feature"`
	Path             string           `json:"path"`
	State            string           `json:"state"`
	WorkflowVersion  int              `json:"workflow_version,omitempty"`
	Phase            string           `json:"phase,omitempty"`
	RequiredSections []string         `json:"required_sections"`
	MissingSections  []string         `json:"missing_sections"`
	HistoricalSpecs  []string         `json:"historical_specs"`
	HistoryIndexes   []string         `json:"history_indexes"`
	PhasePermissions PhasePermissions `json:"phase_permissions"`
}

type PhasePermissions struct {
	SpecAuthoring        bool `json:"spec_authoring"`
	SourceImplementation bool `json:"source_implementation"`
	Delivery             bool `json:"delivery"`
}

type Artifact struct {
	Kind          string   `json:"kind"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Path          string   `json:"path"`
	Version       int      `json:"version"`
	ReadPolicy    string   `json:"read_policy"`
	State         string   `json:"state"`
	Digest        string   `json:"digest,omitempty"`
	SourceRepo    string   `json:"source_repo,omitempty"`
	SourceBranch  string   `json:"source_branch,omitempty"`
	SourceCommit  string   `json:"source_commit,omitempty"`
	SourcePath    string   `json:"source_path,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Applicability []string `json:"applicability,omitempty"`
	Reason        string   `json:"reason"`
}
