package config

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// PluginConfig is the repository's enabled plugin list. Omitted means no plugins.
type PluginConfig struct {
	Enabled []string `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// PlanPluginEnabled changes only plugins.enabled and preserves unrelated YAML.
// An empty list removes the plugins key when that mapping has no other keys.
func PlanPluginEnabled(before []byte, enabled []string) ([]byte, error) {
	for _, id := range enabled {
		if err := validatePluginID(id); err != nil {
			return nil, err
		}
	}
	cfg := Default()
	if err := yaml.Unmarshal(before, cfg); err != nil {
		return nil, fmt.Errorf("parse project config: %w", err)
	}
	if samePluginIDs(cfg.Plugins.Enabled, enabled) {
		return before, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(before, &doc); err != nil {
		return nil, err
	}
	indent := issueDocumentIndent(&doc)
	root, err := documentMapping(&doc, ConfigFileName)
	if err != nil {
		return nil, err
	}
	if root.Anchor != "" {
		return nil, fmt.Errorf("anchored configuration root requires manual review before changing plugins")
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "plugins" && root.Content[i+1].Kind != yaml.MappingNode {
			return nil, fmt.Errorf("plugins must be a mapping")
		}
	}
	if len(enabled) == 0 {
		if plugins := findMapping(root, "plugins"); plugins != nil {
			removeKey(plugins, "enabled")
			if len(plugins.Content) == 0 {
				removeKey(root, "plugins")
			}
		}
	} else {
		setStringSequence(findOrCreateMapping(root, "plugins"), "enabled", enabled)
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(indent)
	if err := encoder.Encode(&doc); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func findMapping(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key && parent.Content[i+1].Kind == yaml.MappingNode {
			return parent.Content[i+1]
		}
	}
	return nil
}

func setStringSequence(parent *yaml.Node, key string, values []string) {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, value := range values {
		seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
	}
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = seq
			return
		}
	}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, seq)
}

func samePluginIDs(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func validatePluginID(id string) error {
	if id == "" {
		return fmt.Errorf("plugin id is empty")
	}
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf("plugin id %q must be lowercase words separated by hyphens", id)
		}
	}
	if id[0] == '-' || id[len(id)-1] == '-' || bytes.Contains([]byte(id), []byte("--")) {
		return fmt.Errorf("plugin id %q must be lowercase words separated by hyphens", id)
	}
	if id[0] >= '0' && id[0] <= '9' {
		return fmt.Errorf("plugin id %q must be lowercase words separated by hyphens", id)
	}
	return nil
}
