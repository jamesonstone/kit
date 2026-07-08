package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Validate checks a document for required sections and other constraints.
func (d *Document) Validate() []ValidationError {
	var errors []ValidationError
	for _, diagnostic := range d.MetadataDiagnostics {
		if diagnostic.Severity != MetadataDiagnosticError {
			continue
		}
		errors = append(errors, ValidationError{
			Document: d.Path,
			Section:  "FRONT MATTER",
			Message:  diagnostic.Message,
			Fix:      diagnostic.Fix,
		})
	}

	required := d.RequiredSections()
	found := make(map[string]bool)
	sections := make(map[string]Section)
	for _, s := range d.Sections {
		key := strings.ToUpper(s.Name)
		found[key] = true
		sections[key] = s
	}

	for _, req := range required {
		key := strings.ToUpper(req)
		if !found[key] {
			errors = append(errors, ValidationError{
				Document: d.Path,
				Section:  req,
				Message:  fmt.Sprintf("missing required section '%s'", req),
				Fix:      fmt.Sprintf("Add a '## %s' section to %s", req, d.Path),
			})
			continue
		}
		if documentTypeRequiresPopulatedSections(d.Type) &&
			!sectionHasVisibleContent(sections[key].Content) {
			errors = append(errors, ValidationError{
				Document: d.Path,
				Section:  req,
				Message:  fmt.Sprintf("required section '%s' is empty", req),
				Fix: fmt.Sprintf(
					"Populate '## %s' in %s or replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`",
					req,
					d.Path,
				),
			})
		}
		if requiresRelationshipSectionValidation(d.Type, key) && !d.FrontMatterPresent {
			section := sections[key]
			if _, err := ParseRelationshipsSection(&section); err != nil {
				errors = append(errors, ValidationError{
					Document: d.Path,
					Section:  req,
					Message:  fmt.Sprintf("invalid relationship syntax: %v", err),
					Fix:      fmt.Sprintf("Use `none` or bullets like `- builds on: 0001-example-feature` in '## %s' in %s", req, d.Path),
				})
			}
		}
	}

	return errors
}

func (d *Document) RequiredSections() []string {
	if d.Type == TypeSpec && d.Metadata != nil && d.Metadata.WorkflowVersion == 2 {
		return SpecV2RequiredSections
	}
	return RequiredSections[d.Type]
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
		return nil
	}

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
