package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PlanIssueConfig changes only known issue fields, keeping unrelated YAML nodes
// and comments. Values must already contain the complete intended target.
func PlanIssueConfig(before []byte, issues IssueConfig) ([]byte, error) {
	if err := issues.Validate(); err != nil {
		return nil, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(before, cfg); err != nil {
		return nil, fmt.Errorf("parse project config: %w", err)
	}
	if cfg.Issues.EffectiveProvider() == issues.EffectiveProvider() && cfg.Issues.Linear == issues.Linear {
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
		return nil, fmt.Errorf("anchored configuration root requires manual review before changing issue settings")
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "<<" {
			return nil, fmt.Errorf("merged configuration root requires manual review before changing issue settings")
		}
		if root.Content[i].Value == "issues" && (hasIssueAlias(root.Content[i]) || hasIssueAlias(root.Content[i+1])) {
			return nil, fmt.Errorf("aliased or anchored issue settings require manual review; no changes made")
		}
	}
	node := findOrCreateMapping(root, "issues")
	setTypedScalar(node, "provider", issues.EffectiveProvider(), "!!str", 0)
	if issues.Linear != cfg.Issues.Linear {
		linear := findOrCreateMapping(node, "linear")
		setTypedScalar(linear, "workspace", issues.Linear.Workspace, "!!str", 0)
		setTypedScalar(linear, "team_id", issues.Linear.TeamID, "!!str", 0)
		if issues.Linear.ProjectID == "" {
			removeKey(linear, "project_id")
		} else {
			setTypedScalar(linear, "project_id", issues.Linear.ProjectID, "!!str", 0)
		}
	}
	return yaml.Marshal(&doc)
}

// ApplyIssueConfig atomically replaces a regular config file only if it still
// matches the bytes previewed by the caller. It never follows a config symlink.
func ApplyIssueConfig(path string, before, after []byte) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("config must be a regular file, not a symlink or special file")
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, before) {
		return fmt.Errorf("configuration changed during planning; rerun the command")
	}
	if bytes.Equal(before, after) {
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kit-config-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err = tmp.Chmod(info.Mode().Perm()); err == nil {
		_, err = tmp.Write(after)
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// Detect a concurrent edit/replacement while the temporary file was written.
	latest, err := os.Lstat(path)
	if err != nil {
		return err
	}
	current, err = os.ReadFile(path)
	if err != nil {
		return err
	}
	if !os.SameFile(info, latest) || !bytes.Equal(current, before) {
		return fmt.Errorf("configuration changed during planning; rerun the command")
	}
	return os.Rename(tmp.Name(), path)
}

func hasIssueAlias(node *yaml.Node) bool {
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return true
	}
	for _, child := range node.Content {
		if hasIssueAlias(child) {
			return true
		}
	}
	return false
}
