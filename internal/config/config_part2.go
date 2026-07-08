package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

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
