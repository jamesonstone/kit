// package config handles .kit.yaml loading and project root discovery.
package config

const ConfigFileName = ".kit.yaml"

const (
	CurrentSchemaVersion              = 1
	InstructionScaffoldVersionVerbose = 1
	InstructionScaffoldVersionTOC     = 2
	InstructionScaffoldVersionMemory  = 3
	DefaultInstructionScaffoldVersion = InstructionScaffoldVersionMemory
	DefaultLoopMaxIterations          = 20
)

// Config represents the .kit.yaml configuration file.
type Config struct {
	SchemaVersion              int                              `yaml:"schema_version"`
	GoalPercentage             int                              `yaml:"goal_percentage"`
	SpecsDir                   string                           `yaml:"specs_dir"`
	SkillsDir                  string                           `yaml:"skills_dir"`
	ConstitutionPath           string                           `yaml:"constitution_path"`
	AllowOutOfOrder            bool                             `yaml:"allow_out_of_order"`
	Loop                       LoopConfig                       `yaml:"loop,omitempty"`
	Agents                     []string                         `yaml:"agents"`
	InstructionScaffoldVersion int                              `yaml:"instruction_scaffold_version"`
	FeatureNaming              FeatureNaming                    `yaml:"feature_naming"`
	FeatureState               map[string]FeatureLifecycleState `yaml:"feature_state,omitempty"`
	RemovedFeatures            []RemovedFeature                 `yaml:"removed_features,omitempty"`
	Prompts                    map[string]map[string]Prompt     `yaml:"prompts,omitempty"`
	Registry                   RegistryConfig                   `yaml:"registry,omitempty"`
	Health                     *HealthConfig                    `yaml:"health,omitempty"`
	Usage                      *UsageConfig                     `yaml:"usage,omitempty"`
	GitHub                     GitHubConfig                     `yaml:"github,omitempty"`
	AWS                        *AWSConfig                       `yaml:"aws,omitempty"`
	ProjectRefresh             ProjectRefreshConfig             `yaml:"project_refresh,omitempty"`
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

type ProjectRefreshConfig struct {
	Constitution ConstitutionRefreshConfig `yaml:"constitution,omitempty"`
}

type ConstitutionRefreshConfig struct {
	FeatureInterval           int    `yaml:"feature_interval,omitempty"`
	MaxAgeDays                int    `yaml:"max_age_days,omitempty"`
	LastReviewedAt            string `yaml:"last_reviewed_at,omitempty"`
	LastCompletedFeatureCount int    `yaml:"last_completed_feature_count,omitempty"`
}

type RegistryConfig struct {
	SchemaVersion int                `yaml:"schema_version,omitempty"`
	Source        RegistrySource     `yaml:"source,omitempty"`
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

type RegistrySource struct {
	Repo   string `yaml:"repo,omitempty"`
	Branch string `yaml:"branch,omitempty"`
}

type RegistryArtifact struct {
	Kind          string                    `yaml:"kind"`
	Slug          string                    `yaml:"slug"`
	Path          string                    `yaml:"path"`
	SourceRepo    string                    `yaml:"source_repo,omitempty"`
	SourceBranch  string                    `yaml:"source_branch,omitempty"`
	SourceCommit  string                    `yaml:"source_commit,omitempty"`
	SourcePath    string                    `yaml:"source_path,omitempty"`
	InstalledHash string                    `yaml:"installed_hash,omitempty"`
	State         string                    `yaml:"state,omitempty"`
	Sections      []RegistryArtifactSection `yaml:"sections,omitempty"`
}

type RegistryArtifactSection struct {
	Key           string `yaml:"key"`
	InstalledHash string `yaml:"installed_hash"`
}

type LoopConfig struct {
	MinConfidence int             `yaml:"min_confidence,omitempty"`
	MaxIterations int             `yaml:"max_iterations,omitempty"`
	Agent         LoopAgentConfig `yaml:"agent,omitempty"`
}

type LoopAgentConfig struct {
	Command string   `yaml:"command,omitempty"`
	Args    []string `yaml:"args,omitempty"`
}

func (c LoopConfig) IsZero() bool {
	if !c.Agent.IsZero() {
		return false
	}
	if c.MinConfidence == 0 && c.MaxIterations == 0 {
		return true
	}
	return c.MinConfidence == 95 && c.MaxIterations == DefaultLoopMaxIterations
}

func (c LoopAgentConfig) IsZero() bool {
	return c.Command == "" && len(c.Args) == 0
}

type FeatureLifecycleState struct {
	Paused bool `yaml:"paused,omitempty"`
}
type RemovedFeature struct {
	Number    int    `yaml:"number"`
	Slug      string `yaml:"slug"`
	DirName   string `yaml:"dir_name"`
	CreatedAt string `yaml:"created_at,omitempty"`
	RemovedAt string `yaml:"removed_at"`
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
		GoalPercentage:             95,
		SpecsDir:                   "docs/specs",
		SkillsDir:                  ".agents/skills",
		ConstitutionPath:           "docs/CONSTITUTION.md",
		AllowOutOfOrder:            false,
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

func (c *Config) IsFeaturePaused(dirName string) bool {
	if c == nil || c.FeatureState == nil {
		return false
	}

	state, ok := c.FeatureState[dirName]
	return ok && state.Paused
}

func (c *Config) SetFeaturePaused(dirName string, paused bool) {
	if paused {
		if c.FeatureState == nil {
			c.FeatureState = make(map[string]FeatureLifecycleState)
		}
		c.FeatureState[dirName] = FeatureLifecycleState{Paused: true}
		return
	}

	if c.FeatureState == nil {
		return
	}

	delete(c.FeatureState, dirName)
	if len(c.FeatureState) == 0 {
		c.FeatureState = nil
	}
}

func (c *Config) RemoveFeatureState(dirName string) {
	c.SetFeaturePaused(dirName, false)
}
func (c *Config) RecordRemovedFeature(record RemovedFeature) {
	if c == nil || record.DirName == "" {
		return
	}

	c.RemoveFeatureState(record.DirName)
	for i := range c.RemovedFeatures {
		if c.RemovedFeatures[i].DirName != record.DirName {
			continue
		}
		c.RemovedFeatures[i] = record
		c.sortRemovedFeatures()
		return
	}

	c.RemovedFeatures = append(c.RemovedFeatures, record)
	c.sortRemovedFeatures()
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
