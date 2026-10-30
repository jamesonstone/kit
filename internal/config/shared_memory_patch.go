package config

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
)

// PlanSharedMemoryConfig updates only the optional shared_memory mapping.
func PlanSharedMemoryConfig(before []byte, desired SharedMemoryConfig) ([]byte, error) {
	if err := desired.Validate(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(before, &cfg); err != nil {
		return nil, err
	}
	if cfg.SharedMemory == desired {
		return before, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(before, &doc); err != nil {
		return nil, err
	}
	root, err := documentMapping(&doc, ConfigFileName)
	if err != nil {
		return nil, err
	}
	if root.Anchor != "" {
		return nil, fmt.Errorf("anchored configuration root requires manual review")
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "<<" || (root.Content[i].Value == "shared_memory" && hasIssueAlias(root.Content[i+1])) {
			return nil, fmt.Errorf("merged or aliased shared-memory configuration requires manual review")
		}
	}
	node := findOrCreateMapping(root, "shared_memory")
	setTypedScalar(node, "enabled", fmt.Sprint(desired.Enabled), "!!bool", 0)
	setTypedScalar(node, "provider", desired.Provider, "!!str", 0)
	setTypedScalar(node, "destination", desired.Destination, "!!str", 0)
	setTypedScalar(node, "guide", desired.Guide, "!!str", 0)
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(issueDocumentIndent(&doc))
	if err := enc.Encode(&doc); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
