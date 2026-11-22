package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PlanGlobalConfig reports the canonical config path and whether a write would
// create the file or add missing default keys. It does not write.
func PlanGlobalConfig(defaults *Config) (path string, create bool, update bool, err error) {
	plan, err := planGlobalConfig(defaults)
	return plan.path, plan.create, plan.update, err
}

// PopulateGlobalConfig creates or fills in the canonical user config.
func PopulateGlobalConfig(defaults *Config) (string, bool, error) {
	plan, err := planGlobalConfig(defaults)
	if err != nil {
		return "", false, err
	}
	if !plan.create && !plan.update {
		return plan.path, false, nil
	}
	dir := filepath.Dir(plan.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", false, fmt.Errorf("failed to create global config directory: %w", err)
	}
	if err := writeYAMLDocument(plan.path, plan.doc); err != nil {
		return "", false, err
	}
	return plan.path, true, nil
}

type globalConfigPlan struct {
	path   string
	doc    *yaml.Node
	create bool
	update bool
}

func planGlobalConfig(defaults *Config) (globalConfigPlan, error) {
	paths, err := globalConfigPaths()
	if err != nil {
		return globalConfigPlan{}, err
	}
	path := paths[0]
	plan := globalConfigPlan{path: path}
	doc, err := readYAMLDocument(path, false)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return plan, err
		}
		plan.create = true
		doc = &yaml.Node{Kind: yaml.DocumentNode}
	}
	plan.doc = doc
	fallback, err := loadGlobalConfigRoot(paths[1:])
	if err != nil {
		return plan, err
	}
	changed, err := appendMissingDefaultConfigFields(doc, defaults, fallback)
	if err != nil {
		return plan, err
	}
	plan.update = changed && !plan.create
	if plan.create || plan.update {
		return plan, checkGlobalConfigWrite(paths)
	}
	return plan, nil
}

func appendMissingDefaultConfigFields(doc *yaml.Node, defaults *Config, fallback *yaml.Node) (bool, error) {
	if defaults == nil {
		defaults = Default()
	}
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
	// Resolve aliases when checking which settings already exist. Write back
	// the original document to preserve its anchors and comments.
	var configured map[string]yaml.Node
	if err := root.Decode(&configured); err != nil {
		return false, fmt.Errorf("failed to inspect existing global config: %w", err)
	}
	changed := false
	for i := 0; i+1 < len(defaultRoot.Content); i += 2 {
		key := defaultRoot.Content[i].Value
		if _, exists := configured[key]; exists || (fallback != nil && mappingHasKey(fallback, key)) {
			continue
		}
		root.Content = append(root.Content, cloneYAMLNode(defaultRoot.Content[i]), cloneYAMLNode(defaultRoot.Content[i+1]))
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
