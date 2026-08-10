package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func PlanUserConfig() (UserConfig, UserConfigDisposition, error) {
	configDir := os.Getenv("KIT_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("resolve user home: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	path := filepath.Join(configDir, "kit", ".kit.yaml")
	info, statErr := os.Lstat(path)
	exists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("inspect user config: %w", statErr)
	}
	if exists && !info.Mode().IsRegular() {
		return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("user config must be a regular file")
	}
	// #nosec G304 -- path is the fixed Kit filename under the explicit config root.
	before, err := os.ReadFile(path)
	if err != nil && (!os.IsNotExist(err) || exists) {
		return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("read user config: %w", err)
	}
	document, err := userConfigDocument(before)
	if err != nil {
		return UserConfig{}, UserConfigDisposition{}, err
	}
	ensureUserDefaults(document)
	after, err := encodeYAML(document)
	if err != nil {
		return UserConfig{}, UserConfigDisposition{}, err
	}
	var config UserConfig
	if err := document.Decode(&config); err != nil {
		return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("decode user config: %w", err)
	}
	if config.SchemaVersion != 1 {
		return UserConfig{}, UserConfigDisposition{}, fmt.Errorf("user config schema_version %d is unsupported", config.SchemaVersion)
	}
	action, state := "none", "preserved"
	if !exists {
		action, state = "create", "planned"
	} else if !bytes.Equal(before, after) {
		action, state = "merge", "planned"
	}
	return config, UserConfigDisposition{
		Path: path, State: state, Action: action, before: string(before),
		after: string(after), exists: exists,
	}, nil
}

func userConfigDocument(content []byte) (*yaml.Node, error) {
	document := &yaml.Node{Kind: yaml.DocumentNode}
	if len(content) == 0 {
		document.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
		return document, nil
	}
	if err := yaml.Unmarshal(content, document); err != nil {
		return nil, fmt.Errorf("parse user config: %w", err)
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("user config must be a YAML mapping")
	}
	return document, nil
}

func ensureUserDefaults(document *yaml.Node) {
	root := document.Content[0]
	ensureScalar(root, "schema_version", "1", "!!int")
	registry := ensureMapping(root, "registry")
	ensureScalar(registry, "repo", "jamesonstone/kit", "!!str")
	ensureScalar(registry, "branch", "main", "!!str")
	ensureScalar(registry, "catalog_path", "registry/catalog.yaml", "!!str")
	bootstrap := ensureMapping(root, "bootstrap")
	ensureScalar(bootstrap, "copy_prompt", "true", "!!bool")
	github := ensureMapping(root, "github")
	if mappingValue(github, "default_assignees") == nil {
		appendMapping(github, "default_assignees", &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"})
	}
}

func ensureMapping(parent *yaml.Node, key string) *yaml.Node {
	if value := mappingValue(parent, key); value != nil {
		if value.Kind != yaml.MappingNode {
			return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		}
		return value
	}
	value := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendMapping(parent, key, value)
	return value
}

func ensureScalar(parent *yaml.Node, key, value, tag string) {
	if mappingValue(parent, key) != nil {
		return
	}
	appendMapping(parent, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
}

func mappingValue(parent *yaml.Node, key string) *yaml.Node {
	for index := 0; index+1 < len(parent.Content); index += 2 {
		if parent.Content[index].Value == key {
			return parent.Content[index+1]
		}
	}
	return nil
}

func appendMapping(parent *yaml.Node, key string, value *yaml.Node) {
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func encodeYAML(document *yaml.Node) ([]byte, error) {
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(document.Content[0]); err != nil {
		return nil, fmt.Errorf("encode user config: %w", err)
	}
	return output.Bytes(), encoder.Close()
}
