package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type mintRepositoryPin struct {
	Kind          string `json:"kind"`
	File          string `json:"file"`
	Ref           string `json:"ref"`
	VersionHint   string `json:"version_hint,omitempty"`
	Compatibility string `json:"compatibility"`
}

var mintModulePin = regexp.MustCompile(`(?m)^\s*(?:(?:require|replace)\s+)?github\.com/jamesonstone/mint[ \t]+([^\r\n]+)`)

func mintCommandTimeout(update bool) time.Duration {
	if update {
		return 10 * time.Minute
	}
	return 30 * time.Second
}

func inspectMintRepositoryPins(root string) ([]mintRepositoryPin, error) {
	pins := []mintRepositoryPin{}
	data, err := readMintPinFile(root, "go.mod")
	if err == nil {
		for _, m := range mintModulePin.FindAllStringSubmatch(string(data), -1) {
			ref, _, _ := strings.Cut(m[1], "//")
			fields := strings.Fields(ref)
			if len(fields) == 0 {
				continue
			}
			kind := "go-module"
			ref = fields[0]
			if strings.Contains(m[1], "=>") {
				kind = "go-module-replacement"
				ref = strings.Join(fields, " ")
			}
			pins = append(pins, mintRepositoryPin{Kind: kind, File: "go.mod", Ref: ref, Compatibility: "not_verified"})
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	dir := filepath.Join(root, ".github", "workflows")
	for _, part := range []string{".github", ".github/workflows"} {
		info, err := os.Lstat(filepath.Join(root, part))
		if os.IsNotExist(err) {
			return pins, nil
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("repository workflow directory must be a real directory")
		}
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return pins, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".yml" && filepath.Ext(entry.Name()) != ".yaml") {
			continue
		}
		file := filepath.ToSlash(filepath.Join(".github", "workflows", entry.Name()))
		data, err := readMintPinFile(root, file)
		if err != nil {
			return nil, err
		}
		var doc yaml.Node
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parse repository workflow %s: %w", file, err)
		}
		var visit func(*yaml.Node)
		visit = func(node *yaml.Node) {
			if node.Kind == yaml.MappingNode {
				for i := 0; i+1 < len(node.Content); i += 2 {
					value := node.Content[i+1]
					if node.Content[i].Value == "uses" && value.Kind == yaml.ScalarNode && strings.HasPrefix(value.Value, "jamesonstone/mint@") {
						pins = append(pins, mintRepositoryPin{Kind: "github-action", File: file, Ref: strings.TrimPrefix(value.Value, "jamesonstone/mint@"), VersionHint: strings.TrimSpace(strings.TrimPrefix(value.LineComment, "#")), Compatibility: "not_verified"})
					}
				}
			}
			for _, child := range node.Content {
				visit(child)
			}
		}
		visit(&doc)
	}
	return pins, nil
}

// Reuse Mint scaffolding's repository path and symlink boundary for reads.
func readMintPinFile(root, relative string) ([]byte, error) {
	return readMintProjectFile(root, relative)
}
