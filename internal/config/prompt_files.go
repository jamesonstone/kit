package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadGlobal() (*Config, bool, error) {
	configPath, err := GlobalConfigPath()
	if err != nil {
		return nil, false, err
	}

	cfg, err := loadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), false, nil
		}
		return nil, false, err
	}

	return cfg, true, nil
}

func loadFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return cfg, nil
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
