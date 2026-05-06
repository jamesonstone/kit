package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PopulateGlobalConfig creates the global config file or appends missing default fields.
func PopulateGlobalConfig(defaults *Config) (string, bool, error) {
	if defaults == nil {
		defaults = Default()
	}

	configPath, err := GlobalConfigPath()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return "", false, fmt.Errorf("failed to create global config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", false, fmt.Errorf("failed to stat %s: %w", configPath, err)
		}
		if err := writeConfigFile(configPath, defaults); err != nil {
			return "", false, err
		}
		return configPath, true, nil
	}

	doc, err := readYAMLDocument(configPath, false)
	if err != nil {
		return "", false, err
	}

	changed, err := appendMissingDefaultConfigFields(doc, defaults)
	if err != nil {
		return "", false, err
	}
	if !changed {
		return configPath, false, nil
	}
	if err := writeYAMLDocument(configPath, doc); err != nil {
		return "", false, err
	}

	return configPath, true, nil
}

func writeConfigFile(configPath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}
	return nil
}

func appendMissingDefaultConfigFields(doc *yaml.Node, defaults *Config) (bool, error) {
	defaultDoc, err := configDocument(defaults)
	if err != nil {
		return false, err
	}

	root, err := documentMapping(doc, "global config")
	if err != nil {
		return false, err
	}
	defaultRoot, err := documentMapping(defaultDoc, "generated config defaults")
	if err != nil {
		return false, err
	}
	changed := false

	for i := 0; i+1 < len(defaultRoot.Content); i += 2 {
		key := defaultRoot.Content[i].Value
		if mappingHasKey(root, key) {
			continue
		}
		root.Content = append(
			root.Content,
			cloneYAMLNode(defaultRoot.Content[i]),
			cloneYAMLNode(defaultRoot.Content[i+1]),
		)
		changed = true
	}

	return changed, nil
}

func configDocument(cfg *Config) (*yaml.Node, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse generated config defaults: %w", err)
	}
	return &doc, nil
}

func mappingHasKey(mapping *yaml.Node, key string) bool {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return true
		}
	}
	return false
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	clone := *node
	if len(node.Content) == 0 {
		return &clone
	}

	clone.Content = make([]*yaml.Node, len(node.Content))
	for i, child := range node.Content {
		clone.Content[i] = cloneYAMLNode(child)
	}
	return &clone
}
