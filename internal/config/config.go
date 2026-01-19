// package config handles .kit.yaml loading and project root discovery.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".kit.yaml"

// Config represents the .kit.yaml configuration file.
type Config struct {
	GoalPercentage   int             `yaml:"goal_percentage"`
	SpecsDir         string          `yaml:"specs_dir"`
	ConstitutionPath string          `yaml:"constitution_path"`
	AllowOutOfOrder  bool            `yaml:"allow_out_of_order"`
	Agents           []string        `yaml:"agents"`
	Branching        BranchingConfig `yaml:"branching"`
	FeatureNaming    FeatureNaming   `yaml:"feature_naming"`
}

// BranchingConfig defines git branching behavior.
type BranchingConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BaseBranch   string `yaml:"base_branch"`
	NameTemplate string `yaml:"name_template"`
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
		ConstitutionPath: "docs/CONSTITUTION.md",
		AllowOutOfOrder:  false,
		Agents:           []string{"AGENTS.md", "CLAUDE.md", "WARP.md"},
		Branching: BranchingConfig{
			Enabled:      true,
			BaseBranch:   "main",
			NameTemplate: "{numeric}-{slug}",
		},
		FeatureNaming: FeatureNaming{
			NumericWidth: 4,
			Separator:    "-",
		},
	}
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

// ConstitutionAbsPath returns the absolute path to the constitution file.
func (c *Config) ConstitutionAbsPath(projectRoot string) string {
	return filepath.Join(projectRoot, c.ConstitutionPath)
}

// ProgressSummaryPath returns the absolute path to PROJECT_PROGRESS_SUMMARY.md.
func (c *Config) ProgressSummaryPath(projectRoot string) string {
	return filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
}
