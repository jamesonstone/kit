package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// RepositoryWorkLedger reads only the repository .kit.yaml work_ledger key.
// It ignores the user file so a user-level target is not written into the repo.
func RepositoryWorkLedger(projectRoot string) (WorkLedgerConfig, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WorkLedgerConfig{}, nil
		}
		return WorkLedgerConfig{}, err
	}
	present, err := yamlKeyPresent(data, "work_ledger")
	if err != nil || !present {
		return WorkLedgerConfig{}, err
	}
	var doc struct {
		WorkLedger WorkLedgerConfig `yaml:"work_ledger"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return WorkLedgerConfig{}, err
	}
	return doc.WorkLedger, nil
}

func LoadWithInspection(projectRoot string) (*Config, Inspection, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		return nil, Inspection{}, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}
	return ParseWithGlobalDefaults(data)
}

// ParseWithGlobalDefaults loads effective settings on top of a captured
// repository file. Use the original bytes as the base for repository writes.
func ParseWithGlobalDefaults(repo []byte) (*Config, Inspection, error) {
	data, err := applyUserConfigOverlay(repo)
	if err != nil {
		return nil, Inspection{}, err
	}
	return ParseWithInspection(data)
}

func applyUserConfigOverlay(repo []byte) ([]byte, error) {
	paths, err := globalConfigPaths()
	if err != nil {
		return nil, err
	}
	global, err := loadGlobalConfigRoot(paths)
	if err != nil {
		return nil, err
	}
	if global == nil {
		return repo, nil
	}
	repoDoc, err := parseYAMLDocument(repo, ConfigFileName)
	if err != nil {
		return nil, err
	}
	// Versions describe the actual repository file and its scaffold. The issue
	// provider and target belong to the repository's delivery workflow.
	removeKey(global, "schema_version")
	removeKey(global, "instruction_scaffold_version")
	removeKey(global, "issues")
	project := documentRoot(repoDoc)
	if usesYAMLInheritance(project) {
		// Decode once with YAML's alias expansion checks before expanding
		// nodes ourselves to retain scalar style.
		var expanded map[string]any
		if err := project.Decode(&expanded); err != nil {
			return nil, fmt.Errorf("failed to resolve repository config: %w", err)
		}
		project, err = resolveConfigYAML(project, make(map[*yaml.Node]bool))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve repository config: %w", err)
		}
	}
	merged := &yaml.Node{Kind: yaml.DocumentNode}
	merged.Content = []*yaml.Node{mergeYAMLNodes(global, project)}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(merged); err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("failed to encode merged Kit config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to encode merged Kit config: %w", err)
	}
	return buf.Bytes(), nil
}

// loadGlobalConfigRoot reads low-priority files first so each higher-priority
// file overrides defaults key by key. Resolve YAML aliases within the file
// where they are declared before combining independent documents.
func loadGlobalConfigRoot(paths []string) (*yaml.Node, error) {
	var merged *yaml.Node
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read global config %s: %w", path, err)
		}
		doc, err := parseYAMLDocument(data, "global config "+path)
		if err != nil {
			return nil, err
		}
		root := documentRoot(doc)
		// Validate the mapping and check alias expansion limits without using
		// the decoded values, which lose scalar spelling and quoting.
		var values map[string]any
		if err := root.Decode(&values); err != nil {
			return nil, fmt.Errorf("failed to parse global config %s: %w", path, err)
		}
		if root.Tag == "!!null" {
			root = documentRoot(nil)
		} else {
			// Decode keys before filtering repository-only fields. A tagged
			// YAML key can have different source text and decoded values.
			root, err = resolveConfigYAML(root, make(map[*yaml.Node]bool))
			if err != nil {
				return nil, fmt.Errorf("failed to resolve global config %s: %w", path, err)
			}
		}
		merged = mergeYAMLNodes(merged, root)
	}
	return merged, nil
}

func usesYAMLInheritance(node *yaml.Node) bool {
	if node.Kind == yaml.AliasNode || node.Tag == "!!merge" {
		return true
	}
	for _, child := range node.Content {
		if usesYAMLInheritance(child) {
			return true
		}
	}
	return false
}

// Resolve merge keys before applying config precedence. Keep scalar syntax
// so checks such as the quoted AWS account ID still see the configured value.
func resolveConfigYAML(node *yaml.Node, aliases map[*yaml.Node]bool) (*yaml.Node, error) {
	resolved := *node
	resolved.Anchor = ""
	switch node.Kind {
	case yaml.AliasNode:
		target := node.Alias
		if target == nil || aliases[target] {
			return nil, fmt.Errorf("invalid or cyclic YAML alias %q", node.Value)
		}
		aliases[target] = true
		result, err := resolveConfigYAML(target, aliases)
		delete(aliases, target)
		return result, err
	case yaml.MappingNode:
		var fields map[string]yaml.Node
		if err := node.Decode(&fields); err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		resolved.Content = nil
		for _, key := range keys {
			field := fields[key]
			value, err := resolveConfigYAML(&field, aliases)
			if err != nil {
				return nil, err
			}
			resolved.Content = append(resolved.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
		}
	case yaml.SequenceNode:
		resolved.Content = nil
		for _, child := range node.Content {
			value, err := resolveConfigYAML(child, aliases)
			if err != nil {
				return nil, err
			}
			resolved.Content = append(resolved.Content, value)
		}
	}
	return &resolved, nil
}

func parseYAMLDocument(data []byte, name string) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}
	return &doc, nil
}

func documentRoot(doc *yaml.Node) *yaml.Node {
	if doc == nil || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}
	return doc.Content[0]
}

func mergeYAMLNodes(base, overlay *yaml.Node) *yaml.Node {
	if overlay == nil {
		return cloneYAMLNode(base)
	}
	if base == nil {
		return cloneYAMLNode(overlay)
	}
	if base.Kind == yaml.MappingNode && overlay.Kind == yaml.MappingNode {
		return mergeYAMLMappings(base, overlay)
	}
	return cloneYAMLNode(overlay)
}

func mergeYAMLMappings(base, overlay *yaml.Node) *yaml.Node {
	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	index := map[string]int{}
	for i := 0; i+1 < len(base.Content); i += 2 {
		key := cloneYAMLNode(base.Content[i])
		out.Content = append(out.Content, key, cloneYAMLNode(base.Content[i+1]))
		index[key.Value] = len(out.Content) - 2
	}
	for i := 0; i+1 < len(overlay.Content); i += 2 {
		key := overlay.Content[i]
		val := overlay.Content[i+1]
		if pos, ok := index[key.Value]; ok {
			out.Content[pos+1] = mergeYAMLNodes(out.Content[pos+1], val)
			continue
		}
		out.Content = append(out.Content, cloneYAMLNode(key), cloneYAMLNode(val))
	}
	return out
}

func yamlKeyPresent(data []byte, key string) (bool, error) {
	doc, err := parseYAMLDocument(data, ConfigFileName)
	if err != nil {
		return false, err
	}
	root := documentRoot(doc)
	if root.Kind != yaml.MappingNode {
		return false, nil
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return true, nil
		}
	}
	return false, nil
}
