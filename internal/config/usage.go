package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfigFile parses a project-shaped Kit config from an explicit path.
// It is used for the global defaults file, which intentionally has no project
// root or schema gate.
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return cfg, nil
}

// UpdateUsageEnabled changes only usage.enabled while preserving all other
// YAML content and comments. Missing global configuration is created safely.
func UpdateUsageEnabled(path string, enabled bool) error {
	var doc *yaml.Node
	if _, err := os.Stat(path); err == nil {
		var readErr error
		doc, readErr = readYAMLDocument(path, false)
		if readErr != nil {
			return readErr
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	} else {
		doc = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}

	root, err := documentMapping(doc, path)
	if err != nil {
		return err
	}
	usage := findOrCreateMapping(root, "usage")
	setTypedScalar(usage, "enabled", fmt.Sprintf("%t", enabled), "!!bool", 0)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return writeYAMLDocument(path, doc)
}
