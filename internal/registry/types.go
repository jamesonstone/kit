package registry

const (
	CatalogSchemaVersion  = 1
	ProjectSchemaVersion  = 2
	ContractSchemaVersion = 1

	KindRuleset  = "ruleset"
	KindWorkflow = "workflow"

	StateManaged     = "managed"
	StateLocalCustom = "local-custom"
	StateConflict    = "conflict"
	StateMissing     = "missing"
)

type SourceConfig struct {
	Repo        string `yaml:"repo,omitempty" json:"repo,omitempty"`
	Branch      string `yaml:"branch,omitempty" json:"branch,omitempty"`
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	CatalogPath string `yaml:"catalog_path,omitempty" json:"catalog_path,omitempty"`
	Revision    string `yaml:"revision,omitempty" json:"revision,omitempty"`
}

type Catalog struct {
	SchemaVersion int               `yaml:"schema_version" json:"schema_version"`
	Artifacts     []CatalogArtifact `yaml:"artifacts" json:"artifacts"`
}

type CatalogArtifact struct {
	Kind         string   `yaml:"kind" json:"kind"`
	Slug         string   `yaml:"slug" json:"slug"`
	Description  string   `yaml:"description" json:"description"`
	Visibility   string   `yaml:"visibility" json:"visibility"`
	SourcePath   string   `yaml:"source_path" json:"source_path"`
	TargetPath   string   `yaml:"target_path" json:"target_path"`
	Version      int      `yaml:"version" json:"version"`
	Digest       string   `yaml:"digest" json:"digest"`
	ReadPolicy   string   `yaml:"read_policy" json:"read_policy"`
	AppliesTo    []string `yaml:"applies_to,omitempty" json:"applies_to,omitempty"`
	Paths        []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
}

type ProjectConfig struct {
	SchemaVersion int                    `yaml:"schema_version" json:"schema_version"`
	Registry      ProjectRegistry        `yaml:"registry" json:"registry"`
	Extra         map[string]interface{} `yaml:",inline" json:"-"`
}

type ProjectRegistry struct {
	SchemaVersion int              `yaml:"schema_version" json:"schema_version"`
	Source        SourceConfig     `yaml:"source" json:"source"`
	Artifacts     []ArtifactRecord `yaml:"artifacts" json:"artifacts"`
}

type ArtifactRecord struct {
	Kind          string        `yaml:"kind" json:"kind"`
	Slug          string        `yaml:"slug" json:"slug"`
	Description   string        `yaml:"description,omitempty" json:"description,omitempty"`
	Path          string        `yaml:"path" json:"path"`
	Version       int           `yaml:"version,omitempty" json:"version,omitempty"`
	ReadPolicy    string        `yaml:"read_policy,omitempty" json:"read_policy,omitempty"`
	AppliesTo     []string      `yaml:"applies_to,omitempty" json:"applies_to,omitempty"`
	Paths         []string      `yaml:"paths,omitempty" json:"paths,omitempty"`
	Dependencies  []string      `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	SourceRepo    string        `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
	SourceBranch  string        `yaml:"source_branch,omitempty" json:"source_branch,omitempty"`
	SourceCommit  string        `yaml:"source_commit,omitempty" json:"source_commit,omitempty"`
	SourcePath    string        `yaml:"source_path,omitempty" json:"source_path,omitempty"`
	InstalledHash string        `yaml:"installed_hash,omitempty" json:"installed_hash,omitempty"`
	ContentHash   string        `yaml:"content_hash,omitempty" json:"content_hash,omitempty"`
	State         string        `yaml:"state" json:"state"`
	Sections      []SectionHash `yaml:"sections,omitempty" json:"sections,omitempty"`
}

type SectionHash struct {
	Key  string `yaml:"key" json:"key"`
	Hash string `yaml:"hash" json:"hash"`
}

type Change struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Before string `json:"-"`
	After  string `json:"-"`
}

type Plan struct {
	State       string         `json:"state"`
	Migration   bool           `json:"migration"`
	Revision    string         `json:"revision,omitempty"`
	Changes     []Change       `json:"changes,omitempty"`
	Artifacts   []ArtifactPlan `json:"artifacts"`
	Diagnostics []string       `json:"diagnostics,omitempty"`
	NextActions []string       `json:"next_actions,omitempty"`
	Config      ProjectConfig  `json:"-"`
}

type ArtifactPlan struct {
	Kind      string   `json:"kind"`
	Slug      string   `json:"slug"`
	Path      string   `json:"path"`
	State     string   `json:"state"`
	Action    string   `json:"action"`
	Conflicts []string `json:"conflicts,omitempty"`
}

func ArtifactKey(kind, slug string) string {
	return kind + "/" + slug
}
