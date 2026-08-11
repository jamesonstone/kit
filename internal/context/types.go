package context

const SchemaVersion = "kit.context/v1"

type Request struct {
	Workflow string   `json:"workflow"`
	Feature  string   `json:"feature,omitempty"`
	Paths    []string `json:"paths,omitempty"`
}

type Contract struct {
	SchemaVersion string             `json:"schema_version"`
	Request       Request            `json:"request"`
	Workflows     []SelectedWorkflow `json:"workflows"`
	Evidence      []EvidenceItem     `json:"evidence"`
	Blocked       bool               `json:"blocked"`
	Diagnostics   []Diagnostic       `json:"diagnostics"`
	NextActions   []string           `json:"next_actions"`
}

type SelectedWorkflow struct {
	Slug         string   `json:"slug"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Digest       string   `json:"digest,omitempty"`
}

type EvidenceItem struct {
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Required bool     `json:"required"`
	State    string   `json:"state"`
	Digest   string   `json:"digest,omitempty"`
	Reasons  []string `json:"reasons"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type WorkflowManifest struct {
	Kind         string             `yaml:"kind"`
	Slug         string             `yaml:"slug"`
	Description  string             `yaml:"description"`
	Dependencies []string           `yaml:"dependencies"`
	Rules        []WorkflowRule     `yaml:"rules"`
	Evidence     []WorkflowEvidence `yaml:"evidence"`
}

type WorkflowRule struct {
	Slug     string `yaml:"slug"`
	Required bool   `yaml:"required"`
}

type WorkflowEvidence struct {
	Kind     string `yaml:"kind"`
	Path     string `yaml:"path"`
	Required bool   `yaml:"required"`
}

type artifactHeader struct {
	Kind string `yaml:"kind"`
	Slug string `yaml:"slug"`
}
