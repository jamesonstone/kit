package document

import (
	"fmt"
	"os"
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

	body := doc.Body
	matches := sectionPattern.FindAllStringSubmatchIndex(body, -1)

	for i, match := range matches {
		name := body[match[2]:match[3]]
		startLine := strings.Count(body[:match[0]], "\n") + block.BodyStartLine

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
