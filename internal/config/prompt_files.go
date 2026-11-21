package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadGlobal() (*Config, bool, error) {
	configPath, err := GlobalConfigPath()
	if err != nil {
		return nil, false, err
	}

	return loadOptionalConfig(configPath)
}

// LoadGlobalUsageEnabled preserves an explicit opt-out stored by older Kit
// versions until the owner sets a value at the canonical path.
func LoadGlobalUsageEnabled() (bool, error) {
	cfg, _, err := LoadGlobal()
	if err != nil {
		return false, err
	}
	if cfg.Usage != nil && cfg.Usage.Enabled != nil {
		return cfg.IsUsageEnabled(), nil
	}

	cfg, err = loadLegacyGlobal()
	if err != nil {
		return false, err
	}
	return cfg.IsUsageEnabled(), nil
}

// LoadGlobalDefaultAssignees falls back to the old file for this setting only.
// An explicitly empty list at the canonical path overrides the fallback.
func LoadGlobalDefaultAssignees() (*[]string, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, found, err := loadOptionalConfig(path)
	if err != nil {
		return nil, err
	}
	if cfg.GitHub.DefaultAssignees != nil {
		return cfg.GitHub.DefaultAssignees, nil
	}
	if found {
		explicit, err := declaresDefaultAssignees(path)
		if explicit || err != nil {
			return nil, err
		}
	}

	cfg, err = loadLegacyGlobal()
	if err != nil {
		return nil, err
	}
	return cfg.GitHub.DefaultAssignees, nil
}

// RepositoryDeclaresDefaultAssignees distinguishes an absent setting from a
// null override, which clears the global default.
func RepositoryDeclaresDefaultAssignees(projectRoot string) (bool, error) {
	return declaresDefaultAssignees(filepath.Join(projectRoot, ConfigFileName))
}

func declaresDefaultAssignees(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	github, present := raw["github"]
	if github == nil {
		return present, nil
	}
	fields, _ := github.(map[string]any)
	_, present = fields["default_assignees"]
	return present, nil
}

func loadLegacyGlobal() (*Config, error) {
	dir, err := LocalDataDir()
	if err != nil {
		return nil, err
	}
	cfg, _, err := loadOptionalConfig(filepath.Join(dir, ConfigFileName))
	return cfg, err
}

func loadOptionalConfig(path string) (*Config, bool, error) {
	cfg, err := LoadConfigFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cfg, true, nil
}

func readYAMLDocument(configPath string, create bool) (*yaml.Node, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) && create {
			data = []byte("{}\n")
		} else {
			return nil, fmt.Errorf("failed to read %s: %w", configPath, err)
		}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		data = []byte("{}\n")
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}
	if doc.Kind == 0 {
		doc.Kind = yaml.DocumentNode
	}
	return &doc, nil
}

func documentMapping(doc *yaml.Node, configPath string) (*yaml.Node, error) {
	if len(doc.Content) == 0 {
		doc.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
		return doc.Content[0], nil
	}
	if doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("failed to update %s: config root must be a YAML mapping", configPath)
	}
	return doc.Content[0], nil
}

func findOrCreateMapping(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		if parent.Content[i+1].Kind != yaml.MappingNode {
			parent.Content[i+1] = &yaml.Node{Kind: yaml.MappingNode}
		}
		return parent.Content[i+1]
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode
}

func removeKey(parent *yaml.Node, key string) {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value != key {
			continue
		}
		parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
		return
	}
}

func writeYAMLDocument(configPath string, doc *yaml.Node) error {
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		_ = encoder.Close()
		return fmt.Errorf("failed to encode %s: %w", configPath, err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to encode %s: %w", configPath, err)
	}

	if err := os.WriteFile(configPath, output.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}
	return nil
}
