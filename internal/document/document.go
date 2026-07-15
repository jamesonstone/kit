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

const (
	WorkflowVersionV2 = 2
	WorkflowVersionV3 = 3
)

// RequiredSections returns the required sections for each document type.
var RequiredSections = map[DocumentType][]string{
	TypeConstitution:    {"PRINCIPLES", "CONSTRAINTS", "NON-GOALS", "DEFINITIONS"},
	TypeBrainstorm:      {"SUMMARY", "USER THESIS", "RELATIONSHIPS", "CODEBASE FINDINGS", "AFFECTED FILES", "DEPENDENCIES", "QUESTIONS", "OPTIONS", "RECOMMENDED STRATEGY", "NEXT STEP"},
	TypeSpec:            {"SUMMARY", "PROBLEM", "GOALS", "NON-GOALS", "USERS", "SKILLS", "RELATIONSHIPS", "DEPENDENCIES", "REQUIREMENTS", "ACCEPTANCE", "EDGE-CASES", "OPEN-QUESTIONS"},
	TypePlan:            {"SUMMARY", "APPROACH", "COMPONENTS", "DATA", "INTERFACES", "DEPENDENCIES", "RISKS", "TESTING"},
	TypeTasks:           {"PROGRESS TABLE", "TASK LIST", "TASK DETAILS", "DEPENDENCIES", "NOTES"},
	TypeAnalysis:        {"UNDERSTANDING", "QUESTIONS", "RESEARCH", "CLARIFICATIONS", "ASSUMPTIONS", "RISKS"},
	TypeProgressSummary: {"FEATURE PROGRESS TABLE", "PROJECT INTENT", "GLOBAL CONSTRAINTS", "FEATURE SUMMARIES", "LAST UPDATED"},
}

var SpecV2RequiredSections = []string{
	"THESIS",
	"CONTEXT",
	"CLARIFICATIONS",
	"REQUIREMENTS",
	"ASSUMPTIONS",
	"ACCEPTANCE CRITERIA",
	"IMPLEMENTATION PLAN",
	"TASK CHECKLIST",
	"VALIDATION MAP",
	"REFLECTION NOTES",
	"DOCUMENTATION UPDATES",
	"DELIVERY DECISION",
	"EVIDENCE",
}

var SpecV3RequiredSections = []string{
	"PURPOSE",
	"CONTEXT",
	"REQUIREMENTS",
	"ACCEPTED PLAN",
	"DECISIONS",
	"DISCOVERIES",
	"VALIDATION",
	"OUTCOME",
	"REPOSITORY MEMORY",
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
	Type                     DocumentType
	Path                     string
	Content                  string
	Body                     string
	FrontMatterRaw           string
	FrontMatterPresent       bool
	Metadata                 *Metadata
	MetadataDiagnostics      []MetadataDiagnostic
	MetadataConflictWarnings []MetadataConflict
	Sections                 []Section
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
	block := splitLeadingFrontMatter(content)
	doc := &Document{
		Type:               docType,
		Path:               path,
		Content:            content,
		Body:               block.Body,
		FrontMatterRaw:     block.Raw,
		FrontMatterPresent: block.Present,
	}
	if doc.Body == "" && content != "" && !block.Present {
		doc.Body = content
	}
	if doc.FrontMatterPresent {
		if block.Err != nil {
			doc.MetadataDiagnostics = append(doc.MetadataDiagnostics, MetadataDiagnostic{
				Severity: MetadataDiagnosticError,
				Field:    "front_matter",
				Message:  block.Err.Error(),
				Fix:      "add a closing `---` delimiter for the YAML front matter block",
			})
		}
		metadata, diagnostics := parseMetadata(doc.FrontMatterRaw, docType)
		doc.Metadata = metadata
		doc.MetadataDiagnostics = append(doc.MetadataDiagnostics, diagnostics...)
	}

	// find all section headers
	body := doc.Body
	matches := sectionPattern.FindAllStringSubmatchIndex(body, -1)

	for i, match := range matches {
		name := body[match[2]:match[3]]
		startLine := strings.Count(body[:match[0]], "\n") + block.BodyStartLine

		// find content between this header and the next (or end)
		contentStart := match[1]
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(body)
		}

		sectionContent := strings.TrimSpace(body[contentStart:contentEnd])

		doc.Sections = append(doc.Sections, Section{
			Name:    name,
			Content: sectionContent,
			Line:    startLine,
		})
	}

	doc.MetadataConflictWarnings = doc.metadataConflicts()

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

	// check required sections
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
		if d.requiresPopulatedSection(key) &&
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
	if d.Type == TypeSpec && d.Metadata != nil {
		switch d.Metadata.WorkflowVersion {
		case WorkflowVersionV2:
			return SpecV2RequiredSections
		case WorkflowVersionV3:
			return SpecV3RequiredSections
		}
	}
	return RequiredSections[d.Type]
}

func (d *Document) IsLivingSpec() bool {
	return d != nil && d.Type == TypeSpec && d.Metadata != nil &&
		(d.Metadata.WorkflowVersion == WorkflowVersionV2 || d.Metadata.WorkflowVersion == WorkflowVersionV3)
}

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
