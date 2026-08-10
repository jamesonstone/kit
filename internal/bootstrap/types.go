package bootstrap

import "github.com/jamesonstone/kit/internal/registry"

const PlanSchemaVersion = 1

type FileDisposition struct {
	Path     string `json:"path"`
	Strategy string `json:"strategy"`
	State    string `json:"state"`
	Action   string `json:"action"`
}

type DirectoryDisposition struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	Action string `json:"action"`
}

type UserConfigDisposition struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	Action string `json:"action"`
	before string
	after  string
	exists bool
}

type exclusiveCreate struct {
	path    string
	content string
	mode    uint32
}

type Plan struct {
	SchemaVersion int                    `json:"schema_version"`
	State         string                 `json:"state"`
	Fresh         bool                   `json:"fresh"`
	Registry      registry.Plan          `json:"registry"`
	Files         []FileDisposition      `json:"bootstrap_files"`
	Directories   []DirectoryDisposition `json:"bootstrap_directories"`
	UserConfig    UserConfigDisposition  `json:"user_config"`
	Diagnostics   []string               `json:"diagnostics,omitempty"`
	NextActions   []string               `json:"next_actions"`
	Prompt        string                 `json:"prompt"`
	root          string
	exclusive     []exclusiveCreate
	userDefaults  UserConfig
}

type UserConfig struct {
	SchemaVersion int `yaml:"schema_version"`
	Registry      struct {
		Repo        string `yaml:"repo"`
		Branch      string `yaml:"branch"`
		CatalogPath string `yaml:"catalog_path"`
	} `yaml:"registry"`
	Bootstrap struct {
		CopyPrompt *bool `yaml:"copy_prompt"`
	} `yaml:"bootstrap"`
	GitHub struct {
		DefaultAssignees []string `yaml:"default_assignees"`
	} `yaml:"github"`
}

func (config UserConfig) Source() registry.SourceConfig {
	return registry.SourceConfig{
		Repo: config.Registry.Repo, Branch: config.Registry.Branch,
		CatalogPath: config.Registry.CatalogPath,
	}
}

func (config UserConfig) CopyPrompt() bool {
	return config.Bootstrap.CopyPrompt == nil || *config.Bootstrap.CopyPrompt
}
