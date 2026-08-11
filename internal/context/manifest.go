package context

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadWorkflow(projectRoot, slug string) (WorkflowManifest, string, []byte, error) {
	if err := validateSlug(slug); err != nil {
		return WorkflowManifest{}, "", nil, err
	}
	relativePath := filepath.ToSlash(filepath.Join("docs", "references", "workflows", slug+".md"))
	_, data, err := secureRead(projectRoot, relativePath)
	if err != nil {
		return WorkflowManifest{}, relativePath, nil, err
	}
	frontMatter, err := markdownFrontMatter(data)
	if err != nil {
		return WorkflowManifest{}, relativePath, data, err
	}
	var manifest WorkflowManifest
	if err := yaml.Unmarshal(frontMatter, &manifest); err != nil {
		return WorkflowManifest{}, relativePath, data, fmt.Errorf("parse workflow front matter: %w", err)
	}
	if manifest.Kind != "workflow" {
		return WorkflowManifest{}, relativePath, data, fmt.Errorf("workflow kind is %q, want workflow", manifest.Kind)
	}
	if manifest.Slug != slug {
		return WorkflowManifest{}, relativePath, data, fmt.Errorf("workflow slug is %q, want %q", manifest.Slug, slug)
	}
	if err := validateManifest(manifest); err != nil {
		return WorkflowManifest{}, relativePath, data, err
	}
	return manifest, relativePath, data, nil
}

func validateManifest(manifest WorkflowManifest) error {
	if strings.TrimSpace(manifest.Description) == "" {
		return fmt.Errorf("workflow description is required")
	}
	seenDependencies := map[string]bool{}
	for _, dependency := range manifest.Dependencies {
		if err := validateSlug(dependency); err != nil {
			return fmt.Errorf("invalid workflow dependency: %w", err)
		}
		if seenDependencies[dependency] {
			return fmt.Errorf("duplicate workflow dependency %q", dependency)
		}
		seenDependencies[dependency] = true
	}
	seenRules := map[string]bool{}
	for _, rule := range manifest.Rules {
		if err := validateSlug(rule.Slug); err != nil {
			return fmt.Errorf("invalid workflow rule: %w", err)
		}
		if seenRules[rule.Slug] {
			return fmt.Errorf("duplicate workflow rule %q", rule.Slug)
		}
		seenRules[rule.Slug] = true
	}
	for _, evidence := range manifest.Evidence {
		if strings.TrimSpace(evidence.Kind) == "" || strings.TrimSpace(evidence.Path) == "" {
			return fmt.Errorf("workflow evidence requires kind and path")
		}
	}
	return nil
}

func validateSlug(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("slug is required")
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '-' {
			continue
		}
		return fmt.Errorf("invalid slug %q", value)
	}
	return nil
}

func markdownFrontMatter(data []byte) ([]byte, error) {
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, []byte("---\n")) {
		return nil, fmt.Errorf("missing workflow front matter")
	}
	remainder := normalized[4:]
	end := bytes.Index(remainder, []byte("\n---\n"))
	if end < 0 {
		return nil, fmt.Errorf("missing closing workflow front matter delimiter")
	}
	return remainder[:end], nil
}

func validateArtifactHeader(data []byte, kind, slug string) error {
	frontMatter, err := markdownFrontMatter(data)
	if err != nil {
		return err
	}
	var header artifactHeader
	if err := yaml.Unmarshal(frontMatter, &header); err != nil {
		return err
	}
	if header.Kind != kind || header.Slug != slug {
		return fmt.Errorf("artifact identity is kind=%q slug=%q, want kind=%q slug=%q", header.Kind, header.Slug, kind, slug)
	}
	return nil
}
