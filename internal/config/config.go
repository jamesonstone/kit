// package config handles .kit.yaml loading and project root discovery.
package config

const ConfigFileName = ".kit.yaml"

const (
	CurrentSchemaVersion              = 2
	InstructionScaffoldVersionVerbose = 1
	InstructionScaffoldVersionTOC     = 2
	InstructionScaffoldVersionMemory  = 3
	DefaultInstructionScaffoldVersion = InstructionScaffoldVersionMemory
)

// Config represents the .kit.yaml configuration file.
type Config struct {
	SchemaVersion              int            `yaml:"schema_version"`
	SpecsDir                   string         `yaml:"specs_dir"`
	ConstitutionPath           string         `yaml:"constitution_path"`
	SourceFileLineLimit        int            `yaml:"source_file_line_limit,omitempty"`
	Agents                     []string       `yaml:"agents"`
	InstructionScaffoldVersion int            `yaml:"instruction_scaffold_version"`
	FeatureNaming              FeatureNaming  `yaml:"feature_naming"`
	Registry                   RegistryConfig `yaml:"registry,omitempty"`
	Health                     *HealthConfig  `yaml:"health,omitempty"`
	Usage                      *UsageConfig   `yaml:"usage,omitempty"`
	GitHub                     GitHubConfig   `yaml:"github,omitempty"`
	AWS                        *AWSConfig     `yaml:"aws,omitempty"`
}

// UsageConfig controls local, private Kit command usage collection. A missing
// value is enabled by default; only an explicit false opts out.
type UsageConfig struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

func (c *Config) IsUsageEnabled() bool {
	return c == nil || c.Usage == nil || c.Usage.Enabled == nil || *c.Usage.Enabled
}

type AWSConfig struct {
	Enabled   *bool  `yaml:"enabled,omitempty"`
	Profile   string `yaml:"profile,omitempty"`
	AccountID string `yaml:"account_id,omitempty"`
	Region    string `yaml:"region,omitempty"`
}

func (c *AWSConfig) IsEnabled() bool {
	if c == nil {
		return false
	}
	return c.Enabled == nil || *c.Enabled
}

func DisabledAWSConfig() *AWSConfig {
	enabled := false
	return &AWSConfig{Enabled: &enabled}
}

type RegistryConfig struct {
	SchemaVersion int                `yaml:"schema_version,omitempty"`
	Artifacts     []RegistryArtifact `yaml:"artifacts,omitempty"`
}

// HealthConfig controls scheduled Kit health maintenance for a project.
// Omitted or null values are managed by default; only explicit false opts out.
type HealthConfig struct {
	Managed *bool `yaml:"managed,omitempty"`
}

func (c *Config) IsHealthManaged() bool {
	return c == nil || c.Health == nil || c.Health.Managed == nil || *c.Health.Managed
}

type GitHubConfig struct {
	Repository       string    `yaml:"repository,omitempty"`
	DefaultBranch    string    `yaml:"default_branch,omitempty"`
	DefaultAssignees *[]string `yaml:"default_assignees,omitempty"`
}

type RegistryArtifact struct {
	Kind          string `yaml:"kind"`
	Slug          string `yaml:"slug"`
	Path          string `yaml:"path"`
	InstalledHash string `yaml:"installed_hash,omitempty"`
	State         string `yaml:"state,omitempty"`
	// Sections is read only from files written by releases that merged local
	// rule edits section by section; its presence marks a file that still
	// carries those edits. Current Kit never writes it.
	Sections []RegistryArtifactSection `yaml:"sections,omitempty"`
}

// RegistryArtifactSection is legacy per-section merge state (see Sections).
type RegistryArtifactSection struct {
	Key           string `yaml:"key"`
	InstalledHash string `yaml:"installed_hash"`
}

// FeatureNaming defines how feature directories are named.
type FeatureNaming struct {
	NumericWidth int    `yaml:"numeric_width"`
	Separator    string `yaml:"separator"`
}

// Default returns a Config with default values per the spec.
func Default() *Config {
	return &Config{
		SchemaVersion:              CurrentSchemaVersion,
		SpecsDir:                   "docs/specs",
		ConstitutionPath:           "docs/CONSTITUTION.md",
		Agents:                     []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"},
		InstructionScaffoldVersion: DefaultInstructionScaffoldVersion,
		FeatureNaming: FeatureNaming{
			NumericWidth: 4,
			Separator:    "-",
		},
	}
}

func IsInstructionScaffoldVersionSupported(version int) bool {
	return version == InstructionScaffoldVersionVerbose ||
		version == InstructionScaffoldVersionTOC ||
		version == InstructionScaffoldVersionMemory
}

func UsesInstructionSupportDocs(version int) bool {
	return version == InstructionScaffoldVersionTOC || version == InstructionScaffoldVersionMemory
}

func (c *Config) EffectiveInstructionScaffoldVersion() int {
	if c == nil || !IsInstructionScaffoldVersionSupported(c.InstructionScaffoldVersion) {
		return DefaultInstructionScaffoldVersion
	}

	return c.InstructionScaffoldVersion
}

func (c *Config) RegistryArtifact(kind, slug string) (RegistryArtifact, bool) {
	if c == nil {
		return RegistryArtifact{}, false
	}
	for _, artifact := range c.Registry.Artifacts {
		if artifact.Kind == kind && artifact.Slug == slug {
			return artifact, true
		}
	}
	return RegistryArtifact{}, false
}
