package plugin

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type manifest struct {
	SchemaVersion int    `yaml:"schema_version"`
	ID            string `yaml:"id"`
	Instructions  string `yaml:"instructions"`
	Config        string `yaml:"config"`
}

// LoadEnabled reads plugins in list order. An empty list reads nothing.
// A declared config file is not opened; its project-relative path is returned
// on Block.ConfigPath so the agent can read it.
func LoadEnabled(projectRoot string, ids []string) ([]Block, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := map[string]bool{}
	blocks := make([]Block, 0, len(ids))
	for _, id := range ids {
		if err := ValidateID(id); err != nil {
			return nil, err
		}
		if seen[id] {
			return nil, fmt.Errorf("plugin %s is listed more than once", id)
		}
		seen[id] = true
		block, err := loadOne(projectRoot, id)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func loadOne(projectRoot, id string) (Block, error) {
	dir := filepath.Join(RootDir, id)
	manifestPath := filepath.Join(dir, ManifestName)
	data, err := readProjectFile(projectRoot, manifestPath, maxManifestBytes)
	if errors.Is(err, os.ErrNotExist) {
		return Block{}, fmt.Errorf("plugin %s was not found at %s", id, filepath.ToSlash(manifestPath))
	}
	if err != nil {
		return Block{}, fmt.Errorf("plugin %s: %w", id, err)
	}
	var doc manifest
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Block{}, fmt.Errorf("plugin %s: parse %s: %w", id, ManifestName, err)
	}
	if doc.SchemaVersion != SchemaVersion {
		return Block{}, fmt.Errorf("plugin %s %s schema_version must be %d", id, ManifestName, SchemaVersion)
	}
	if doc.ID != id {
		return Block{}, fmt.Errorf("plugin %s %s id is %q", id, ManifestName, doc.ID)
	}
	instructions, err := pluginRelative(doc.Instructions)
	if err != nil {
		return Block{}, fmt.Errorf("plugin %s instructions: %w", id, err)
	}
	body, err := readProjectFile(projectRoot, filepath.Join(dir, filepath.FromSlash(instructions)), maxInstructionBytes)
	if err != nil {
		return Block{}, fmt.Errorf("plugin %s instructions: %w", id, err)
	}
	text, err := instructionText(id, body)
	if err != nil {
		return Block{}, err
	}
	block := Block{ID: id, Body: text}
	if strings.TrimSpace(doc.Config) == "" {
		return block, nil
	}
	configPath, err := pluginRelative(doc.Config)
	if err != nil {
		return Block{}, fmt.Errorf("plugin %s config: %w", id, err)
	}
	block.ConfigPath = path.Join(RootDir, id, configPath)
	return block, nil
}

func instructionText(id string, raw []byte) (string, error) {
	if !utf8.Valid(raw) {
		return "", fmt.Errorf("plugin %s instructions must be UTF-8", id)
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return "", fmt.Errorf("plugin %s instructions are empty", id)
	}
	if strings.Contains(text, "\r") {
		return "", fmt.Errorf("plugin %s instructions must use LF line endings", id)
	}
	if strings.Contains(text, "KIT-PLUGIN") || strings.Contains(text, "KIT-MANAGED CONTRACT") {
		return "", fmt.Errorf("plugin %s instructions contain a managed marker", id)
	}
	return text, nil
}

func pluginRelative(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("path is required")
	}
	slash := filepath.ToSlash(value)
	if filepath.IsAbs(value) || strings.HasPrefix(slash, "/") || strings.Contains(slash, `\`) || strings.Contains(slash, "..") {
		return "", fmt.Errorf("path %q must stay inside the plugin directory", value)
	}
	cleaned := path.Clean(slash)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, ":") {
		return "", fmt.Errorf("path %q must stay inside the plugin directory", value)
	}
	if !filepath.IsLocal(filepath.FromSlash(cleaned)) {
		return "", fmt.Errorf("path %q must stay inside the plugin directory", value)
	}
	return cleaned, nil
}

func readProjectFile(projectRoot, rel string, limit int) ([]byte, error) {
	rel = filepath.Clean(rel)
	if !filepath.IsLocal(rel) {
		return nil, fmt.Errorf("path %s escapes the project", filepath.ToSlash(rel))
	}
	root, err := os.OpenRoot(projectRoot)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	current := ""
	parts := strings.Split(rel, string(filepath.Separator))
	for i, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("path %s escapes the project", filepath.ToSlash(rel))
		}
		current = filepath.Join(current, part)
		info, err := root.Lstat(current)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("%s is a symlink", filepath.ToSlash(current))
		}
		last := i == len(parts)-1
		if !last && !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", filepath.ToSlash(current))
		}
		if last && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%s is not a regular file", filepath.ToSlash(current))
		}
	}
	data, err := root.ReadFile(rel)
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes", filepath.ToSlash(rel), limit)
	}
	return data, nil
}
