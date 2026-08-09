package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const preambleSection = "__preamble__"

type DocumentMetadata struct {
	Kind              string   `yaml:"kind"`
	Slug              string   `yaml:"slug"`
	Description       string   `yaml:"description"`
	Status            string   `yaml:"status"`
	RegistryScope     string   `yaml:"registry_scope"`
	ReadPolicyDefault string   `yaml:"read_policy_default"`
	AppliesTo         []string `yaml:"applies_to"`
	Paths             []string `yaml:"paths"`
	Dependencies      []string `yaml:"dependencies"`
}

type MarkdownDocument struct {
	Metadata DocumentMetadata
	Order    []string
	Sections map[string]string
}

func ParseMarkdown(content string) (MarkdownDocument, error) {
	normalized := normalizeContent(content)
	frontMatter, body, err := splitFrontMatter(normalized)
	if err != nil {
		return MarkdownDocument{}, err
	}
	var metadata DocumentMetadata
	if err := yaml.Unmarshal([]byte(frontMatter), &metadata); err != nil {
		return MarkdownDocument{}, fmt.Errorf("parse artifact front matter: %w", err)
	}
	doc := MarkdownDocument{Metadata: metadata, Sections: map[string]string{}}
	doc.Order, doc.Sections = splitSections("---\n" + frontMatter + "\n---\n" + body)
	return doc, nil
}

func ValidateDocument(doc MarkdownDocument, artifact CatalogArtifact) error {
	if doc.Metadata.Kind != artifact.Kind {
		return fmt.Errorf("front matter kind %q does not match catalog kind %q", doc.Metadata.Kind, artifact.Kind)
	}
	if doc.Metadata.Slug != artifact.Slug {
		return fmt.Errorf("front matter slug %q does not match catalog slug %q", doc.Metadata.Slug, artifact.Slug)
	}
	if strings.TrimSpace(doc.Metadata.Description) == "" {
		return fmt.Errorf("front matter description is required")
	}
	return nil
}

func HashContent(content string) string {
	sum := sha256.Sum256([]byte(normalizeContent(content)))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func HashSections(content string) ([]SectionHash, error) {
	doc, err := ParseMarkdown(content)
	if err != nil {
		return nil, err
	}
	result := make([]SectionHash, 0, len(doc.Sections))
	for key, section := range doc.Sections {
		result = append(result, SectionHash{Key: key, Hash: HashContent(section)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result, nil
}

func splitFrontMatter(content string) (string, string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return "", "", fmt.Errorf("artifact must start with YAML front matter")
	}
	rest := strings.TrimPrefix(content, "---\n")
	index := strings.Index(rest, "\n---\n")
	if index < 0 {
		return "", "", fmt.Errorf("artifact front matter is not closed")
	}
	return rest[:index], rest[index+5:], nil
}

func splitSections(content string) ([]string, map[string]string) {
	lines := strings.Split(normalizeContent(content), "\n")
	sections := map[string]string{}
	var order []string
	current := preambleSection
	var buffer []string
	inFence := false
	flush := func() {
		if len(buffer) == 0 {
			return
		}
		if _, exists := sections[current]; !exists {
			order = append(order, current)
		}
		sections[current] = strings.TrimRight(strings.Join(buffer, "\n"), "\n") + "\n"
		buffer = nil
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
		}
		if !inFence && strings.HasPrefix(line, "## ") {
			flush()
			current = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
		buffer = append(buffer, line)
	}
	flush()
	return order, sections
}

func renderSections(order []string, sections map[string]string) string {
	var builder strings.Builder
	written := map[string]bool{}
	for _, key := range order {
		section, ok := sections[key]
		if !ok || written[key] {
			continue
		}
		builder.WriteString(section)
		written[key] = true
	}
	var remaining []string
	for key := range sections {
		if !written[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	for _, key := range remaining {
		builder.WriteString(sections[key])
	}
	return normalizeContent(builder.String())
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.TrimRight(content, "\n") + "\n"
}
