// package document handles markdown document parsing and validation.
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DocumentType represents a type of Kit document.
type DocumentType string

const (
	TypeConstitution    DocumentType = "CONSTITUTION"
	TypeBrainstorm      DocumentType = "BRAINSTORM"
	TypeSpec            DocumentType = "SPEC"
	TypePlan            DocumentType = "PLAN"
	TypeTasks           DocumentType = "TASKS"
	TypeAnalysis        DocumentType = "ANALYSIS"
	TypeProgressSummary DocumentType = "PROJECT_PROGRESS_SUMMARY"
)

// RequiredSections returns the required sections for each document type.
var RequiredSections = map[DocumentType][]string{
	TypeConstitution:    {"PRINCIPLES", "CONSTRAINTS", "NON-GOALS", "DEFINITIONS"},
	TypeBrainstorm:      {"SUMMARY", "USER THESIS", "CODEBASE FINDINGS", "AFFECTED FILES", "DEPENDENCIES", "QUESTIONS", "OPTIONS", "RECOMMENDED STRATEGY", "NEXT STEP"},
	TypeSpec:            {"SUMMARY", "PROBLEM", "GOALS", "NON-GOALS", "USERS", "SKILLS", "DEPENDENCIES", "REQUIREMENTS", "ACCEPTANCE", "EDGE-CASES", "OPEN-QUESTIONS"},
	TypePlan:            {"SUMMARY", "APPROACH", "COMPONENTS", "DATA", "INTERFACES", "DEPENDENCIES", "RISKS", "TESTING"},
	TypeTasks:           {"PROGRESS TABLE", "TASK LIST", "TASK DETAILS", "DEPENDENCIES", "NOTES"},
	TypeAnalysis:        {"UNDERSTANDING", "QUESTIONS", "RESEARCH", "CLARIFICATIONS", "ASSUMPTIONS", "RISKS"},
	TypeProgressSummary: {"FEATURE PROGRESS TABLE", "PROJECT INTENT", "GLOBAL CONSTRAINTS", "FEATURE SUMMARIES", "LAST UPDATED"},
}

var (
	// sectionPattern matches markdown headers like "## SECTION NAME"
	sectionPattern = regexp.MustCompile(`(?m)^##\s+(.+)$`)
	// placeholderPattern matches TODO comments
	placeholderPattern = regexp.MustCompile(`<!--\s*TODO:.*?-->`)
	// linkPattern matches traceability links like [SPEC-01] or [PLAN-01]
	linkPattern = regexp.MustCompile(`\[(?:SPEC|PLAN)-\d+\]`)
)

// Section represents a parsed section from a document.
type Section struct {
	Name    string
	Content string
	Line    int
}

// Document represents a parsed Kit document.
type Document struct {
	Type     DocumentType
	Path     string
	Content  string
	Sections []Section
}

// ParseFile reads and parses a document from the filesystem.
func ParseFile(path string, docType DocumentType) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(string(content), path, docType), nil
}

// Parse parses a document from its content.
func Parse(content string, path string, docType DocumentType) *Document {
	doc := &Document{
		Type:    docType,
		Path:    path,
		Content: content,
	}

	// find all section headers
	matches := sectionPattern.FindAllStringSubmatchIndex(content, -1)
	lines := strings.Split(content, "\n")

	for i, match := range matches {
		name := content[match[2]:match[3]]
		startLine := strings.Count(content[:match[0]], "\n") + 1

		// find content between this header and the next (or end)
		contentStart := match[1]
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(content)
		}

		sectionContent := strings.TrimSpace(content[contentStart:contentEnd])

		doc.Sections = append(doc.Sections, Section{
			Name:    name,
			Content: sectionContent,
			Line:    startLine,
		})
	}

	// for line counting
	_ = lines

	return doc
}

// ValidationError represents a validation error with context.
type ValidationError struct {
	Document string
	Section  string
	Message  string
	Fix      string
}

func (e ValidationError) Error() string {
	if e.Fix != "" {
		return fmt.Sprintf("%s: %s. %s", e.Document, e.Message, e.Fix)
	}
	return fmt.Sprintf("%s: %s", e.Document, e.Message)
}

// Validate checks a document for required sections and other constraints.
func (d *Document) Validate() []ValidationError {
	var errors []ValidationError

	// check required sections
	required := RequiredSections[d.Type]
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
	}

	return errors
}

// HasUnresolvedPlaceholders checks if the document has TODO placeholders.
func (d *Document) HasUnresolvedPlaceholders() bool {
	return placeholderPattern.MatchString(d.Content)
}

// GetUnresolvedPlaceholders returns all unresolved placeholders.
func (d *Document) GetUnresolvedPlaceholders() []string {
	return placeholderPattern.FindAllString(d.Content, -1)
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
	return linkPattern.FindAllString(d.Content, -1)
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
	if len(text) > 120 {
		text = text[:117] + "..."
	}
	return text
}
