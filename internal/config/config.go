// package config handles .kit.yaml loading and project root discovery.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".kit.yaml"

const (
	InstructionScaffoldVersionVerbose = 1
	InstructionScaffoldVersionTOC     = 2
	DefaultInstructionScaffoldVersion = InstructionScaffoldVersionTOC
)

// Config represents the .kit.yaml configuration file.
type Config struct {
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
}

type RegistryConfig struct {
	SchemaVersion int                `yaml:"schema_version,omitempty"`
	Source        RegistrySource     `yaml:"source,omitempty"`
	Artifacts     []RegistryArtifact `yaml:"artifacts,omitempty"`
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
	return c.MinConfidence == 95 && c.MaxIterations == 20
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
		GoalPercentage:   95,
		SpecsDir:         "docs/specs",
		SkillsDir:        ".agents/skills",
		ConstitutionPath: "docs/CONSTITUTION.md",
		AllowOutOfOrder:  false,
		Loop: LoopConfig{
			MinConfidence: 95,
			MaxIterations: 20,
		},
		Agents: []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"},
		FeatureNaming: FeatureNaming{
			NumericWidth: 4,
			Separator:    "-",
		},
	}
}

func IsInstructionScaffoldVersionSupported(version int) bool {
	return version == InstructionScaffoldVersionVerbose || version == InstructionScaffoldVersionTOC
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

func (c *Config) UpsertRegistryArtifact(artifact RegistryArtifact) {
	if c == nil || artifact.Kind == "" || artifact.Slug == "" {
		return
	}
	if c.Registry.SchemaVersion == 0 {
		c.Registry.SchemaVersion = 1
	}
	for i := range c.Registry.Artifacts {
		if c.Registry.Artifacts[i].Kind == artifact.Kind && c.Registry.Artifacts[i].Slug == artifact.Slug {
			c.Registry.Artifacts[i] = artifact
			c.sortRegistryArtifacts()
			return
		}
	}
	c.Registry.Artifacts = append(c.Registry.Artifacts, artifact)
	c.sortRegistryArtifacts()
}

func (c *Config) sortRegistryArtifacts() {
	sort.SliceStable(c.Registry.Artifacts, func(i, j int) bool {
		if c.Registry.Artifacts[i].Kind != c.Registry.Artifacts[j].Kind {
			return c.Registry.Artifacts[i].Kind < c.Registry.Artifacts[j].Kind
		}
		return c.Registry.Artifacts[i].Slug < c.Registry.Artifacts[j].Slug
	})
}

func (c *Config) sortRemovedFeatures() {
	sort.SliceStable(c.RemovedFeatures, func(i, j int) bool {
		if c.RemovedFeatures[i].Number != c.RemovedFeatures[j].Number {
			return c.RemovedFeatures[i].Number < c.RemovedFeatures[j].Number
		}
		return c.RemovedFeatures[i].DirName < c.RemovedFeatures[j].DirName
	})
}

// FindProjectRoot traverses upward from the current directory to find .kit.yaml.
// Returns the directory containing .kit.yaml, or an error if not found.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// reached filesystem root
			return "", fmt.Errorf("%s not found. Run 'kit init' to initialize a project", ConfigFileName)
		}
		dir = parent
	}
}

// Load reads and parses the .kit.yaml from the given project root.
func Load(projectRoot string) (*Config, error) {
	configPath := filepath.Join(projectRoot, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ConfigFileName, err)
	}

	return cfg, nil
}

// LoadOrDefault attempts to load config from project root, returns default if not found.
func LoadOrDefault(projectRoot string) *Config {
	cfg, err := Load(projectRoot)
	if err != nil {
		return Default()
	}
	return cfg
}

// Save writes the config to .kit.yaml in the given project root.
func Save(projectRoot string, cfg *Config) error {
	configPath := filepath.Join(projectRoot, ConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ConfigFileName, err)
	}

	return nil
}

// Exists checks if .kit.yaml exists in the given directory.
func Exists(dir string) bool {
	configPath := filepath.Join(dir, ConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// SpecsPath returns the absolute path to the specs directory.
func (c *Config) SpecsPath(projectRoot string) string {
	return filepath.Join(projectRoot, c.SpecsDir)
}

// SkillsPath returns the absolute path to the skills directory.
func (c *Config) SkillsPath(projectRoot string) string {
	return filepath.Join(projectRoot, c.SkillsDir)
}

// ConstitutionAbsPath returns the absolute path to the constitution file.
func (c *Config) ConstitutionAbsPath(projectRoot string) string {
	return filepath.Join(projectRoot, c.ConstitutionPath)
}

// ProgressSummaryPath returns the absolute path to PROJECT_PROGRESS_SUMMARY.md.
func (c *Config) ProgressSummaryPath(projectRoot string) string {
	return filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
}
