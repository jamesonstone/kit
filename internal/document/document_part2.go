package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// requiresPopulatedSection reports whether a section must contain visible
// content for the document's current workflow and phase.
func (d *Document) requiresPopulatedSection(section string) bool {
	if d.Type != TypeSpec || d.Metadata == nil || d.Metadata.WorkflowVersion != WorkflowVersionV3 {
		return documentTypeRequiresPopulatedSections(d.Type)
	}

	section = strings.ToUpper(strings.TrimSpace(section))
	switch strings.ToLower(strings.TrimSpace(d.Metadata.Phase)) {
	case "ready", "implement", "validate", "reflect", "deliver", "complete", "blocked":
		for _, required := range []string{"PURPOSE", "CONTEXT", "REQUIREMENTS", "ACCEPTED PLAN"} {
			if section == required {
				return true
			}
		}
	}

	switch strings.ToLower(strings.TrimSpace(d.Metadata.Phase)) {
	case "deliver", "complete":
		return true
	default:
		return false
	}
}

// RequiresPopulatedSection reports whether the document's current workflow
// phase requires visible content in section.
func (d *Document) RequiresPopulatedSection(section string) bool {
	return d.requiresPopulatedSection(strings.ToUpper(strings.TrimSpace(section)))
}

// HasUnresolvedPlaceholders checks if the document has TODO placeholders.
func (d *Document) HasUnresolvedPlaceholders() bool {
	return placeholderPattern.MatchString(d.Body)
}

// GetUnresolvedPlaceholders returns all unresolved placeholders.
func (d *Document) GetUnresolvedPlaceholders() []string {
	return placeholderPattern.FindAllString(d.Body, -1)
}

// GetSection returns a section by name (case-insensitive).
func (d *Document) GetSection(name string) *Section {
	name = strings.ToUpper(name)
	for _, s := range d.Sections {
		if strings.ToUpper(s.Name) == name {
			return &s
		}
	}
	return nil
}

// HasSection checks if a section exists.
func (d *Document) HasSection(name string) bool {
	return d.GetSection(name) != nil
}

// GetLinks returns all traceability links in the document.
func (d *Document) GetLinks() []string {
	return linkPattern.FindAllString(d.Body, -1)
}

// Exists checks if a document file exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Write writes content to a document file, creating parent directories if needed.
func Write(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteIfNotExists writes content only if the file doesn't exist.
func WriteIfNotExists(path string, content string) (created bool, err error) {
	if Exists(path) {
		return false, nil
	}
	return true, Write(path, content)
}

// MergeDocument merges new content into an existing document, preserving existing sections.
// This adds any missing required sections from the template.
func MergeDocument(existingPath string, templateContent string, docType DocumentType) error {
	existing, err := ParseFile(existingPath, docType)
	if err != nil {
		// file doesn't exist, just write the template
		return Write(existingPath, templateContent)
	}

	template := Parse(templateContent, "", docType)

	// find sections in template that are missing from existing
	var missingSections []Section
	for _, ts := range template.Sections {
		if !existing.HasSection(ts.Name) {
			missingSections = append(missingSections, ts)
		}
	}

	if len(missingSections) == 0 {
		return nil // nothing to merge
	}

	// append missing sections to existing content
	content := existing.Content
	for _, s := range missingSections {
		content += fmt.Sprintf("\n\n## %s\n\n%s", s.Name, s.Content)
	}

	return Write(existingPath, content)
}

// ExtractFirstParagraph extracts the first non-empty paragraph after a section header.
func ExtractFirstParagraph(section *Section) string {
	if section == nil {
		return ""
	}

	lines := strings.Split(section.Content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(result) > 0 {
				break
			}
			continue
		}
		line = visibleLineContent(trimmed)
		if line == "" {
			continue
		}
		result = append(result, line)
	}

	text := strings.Join(result, " ")
	if isExplicitSectionFallbackText(text) {
		return ""
	}
	return text
}
