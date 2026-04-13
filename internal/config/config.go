// package config handles .kit.yaml loading and project root discovery.
package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	Agents                     []string                         `yaml:"agents"`
	InstructionScaffoldVersion int                              `yaml:"instruction_scaffold_version"`
	FeatureNaming              FeatureNaming                    `yaml:"feature_naming"`
	FeatureState               map[string]FeatureLifecycleState `yaml:"feature_state,omitempty"`
}

type FeatureLifecycleState struct {
	Paused bool `yaml:"paused,omitempty"`
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
		Agents:           []string{"AGENTS.md", "CLAUDE.md"},
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
