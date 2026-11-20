package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UserKitConfigFile is the optional user-level Kit config. It is not committed
// and it is not ~/.config/kit/.kit.yaml. Repository .kit.yaml overrides it
// key by key.
const UserKitConfigFile = ".kit/.kit.yaml"

// UserKitConfigPath returns $HOME/.kit/.kit.yaml. Kit never creates the file.
func UserKitConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("failed to resolve HOME for the user Kit config")
	}
	return filepath.Join(home, filepath.FromSlash(UserKitConfigFile)), nil
}

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

func readMergedConfig(projectRoot string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}
	return applyUserConfigOverlay(data)
}

func applyUserConfigOverlay(repo []byte) ([]byte, error) {
	path, err := UserKitConfigPath()
	if err != nil {
		return nil, err
	}
	user, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repo, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", UserKitConfigFile, err)
	}
	if len(bytes.TrimSpace(user)) == 0 {
		return repo, nil
	}
	merged, err := overlayConfigYAML(user, repo)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// overlayConfigYAML deep-merges mappings. A repository scalar or sequence
// replaces the user value for that key. Absent repository keys keep the user value.
func overlayConfigYAML(user, repo []byte) ([]byte, error) {
	userDoc, err := parseYAMLDocument(user, "user Kit config")
	if err != nil {
		return nil, err
	}
	repoDoc, err := parseYAMLDocument(repo, ConfigFileName)
	if err != nil {
		return nil, err
	}
	merged := &yaml.Node{Kind: yaml.DocumentNode}
	merged.Content = []*yaml.Node{mergeYAMLNodes(documentRoot(userDoc), documentRoot(repoDoc))}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(merged); err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("failed to encode merged Kit config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
