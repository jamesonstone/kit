package cli

import (
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func expectedSectionsFor(docType document.DocumentType) []string {
	template := document.Parse(templateContentFor(docType), "", docType)
	sections := make([]string, 0, len(template.Sections))
	for _, section := range template.Sections {
		sections = append(sections, section.Name)
	}
	return sections
}

func templateContentFor(docType document.DocumentType) string {
	switch docType {
	case document.TypeConstitution:
		return templates.Constitution
	case document.TypeBrainstorm:
		return templates.BrainstormArtifact
	case document.TypeSpec:
		return templates.Spec
	case document.TypePlan:
		return templates.Plan
	case document.TypeTasks:
		return templates.Tasks
	case document.TypeProgressSummary:
		return templates.ProjectProgressSummary
	default:
		return ""
	}
}

func tableExpectationsFor(docType document.DocumentType) []tableExpectation {
	switch docType {
	case document.TypeBrainstorm, document.TypePlan:
		return []tableExpectation{
			{
				Section:     "DEPENDENCIES",
				Headers:     []string{"Dependency", "Type", "Location", "Used For", "Status"},
				RequireRows: true,
			},
		}
	case document.TypeSpec:
		return []tableExpectation{
			{
				Section:     "SKILLS",
				Headers:     []string{"SKILL", "SOURCE", "PATH", "TRIGGER", "REQUIRED"},
				RequireRows: true,
			},
			{
				Section:     "DEPENDENCIES",
				Headers:     []string{"Dependency", "Type", "Location", "Used For", "Status"},
				RequireRows: true,
			},
		}
	case document.TypeTasks:
		return []tableExpectation{
			{
				Section:     "PROGRESS TABLE",
				Headers:     []string{"ID", "TASK", "STATUS", "OWNER", "DEPENDENCIES"},
				RequireRows: true,
			},
		}
	default:
		return nil
	}
}

type tableExpectation struct {
	Section     string
	Headers     []string
	RequireRows bool
}

func validateTableSection(content string, headers []string, requireRows bool) string {
	rows := tableRows(content)
	if len(rows) < 2 {
		return "missing table header or divider"
	}
	if !slicesEqual(rows[0], headers) {
		return "unexpected headers"
	}
	if requireRows && len(rows) < 3 {
		return "missing data rows"
	}
	return ""
}

func tableRows(content string) [][]string {
	var rows [][]string
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || !strings.HasPrefix(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		cells := make([]string, 0, len(parts)-2)
		for _, cell := range parts[1 : len(parts)-1] {
			cells = append(cells, strings.TrimSpace(cell))
		}
		rows = append(rows, cells)
	}
	return rows
}

func progressTableTaskIDs(section *document.Section) ([]string, bool) {
	if section == nil {
		return nil, false
	}

	rows := tableRows(section.Content)
	if len(rows) < 3 {
		return nil, false
	}

	var ids []string
	for _, row := range rows[2:] {
		if len(row) == 0 || row[0] == "" {
			continue
		}
		ids = append(ids, row[0])
	}
	return ids, true
}

func matchesToSet(matches [][]string) map[string]bool {
	set := make(map[string]bool, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			set[match[1]] = true
		}
	}
	return set
}

func meaningfulSectionContent(content string) bool {
	cleaned := reconcileCommentPattern.ReplaceAllString(content, "")
	return strings.TrimSpace(cleaned) != ""
}

func slicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
