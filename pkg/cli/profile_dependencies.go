package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

const (
	frontendProfileDependencyName     = "Frontend profile"
	frontendProfileDependencyType     = "profile"
	frontendProfileDependencyLocation = "--profile=frontend"
	designMaterialsDependencyName     = "Design materials"
	designMaterialsDependencyType     = "design"
)

type profileDependencyRow struct {
	Dependency string
	Type       string
	Location   string
	UsedFor    string
	Status     string
}

func designMaterialsRelPath(featureDirName string) string {
	return filepath.ToSlash(filepath.Join("docs", "notes", featureDirName, "design"))
}

func featureHasActiveFrontendProfileDependency(featurePath string) bool {
	for _, source := range frontendProfileDependencySources(featurePath) {
		if !document.Exists(source.path) {
			continue
		}
		content, err := os.ReadFile(source.path)
		if err != nil {
			continue
		}
		doc := document.Parse(string(content), source.path, source.docType)
		if hasActiveFrontendProfileDependency(doc.GetSection("DEPENDENCIES")) {
			return true
		}
	}
	return false
}

func ensureFrontendProfileDependencyRows(docPath string, docType document.DocumentType, featureDirName string) (bool, error) {
	content, err := os.ReadFile(docPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", docPath, err)
	}

	updated, changed := appendFrontendProfileDependencyRows(string(content), docType, featureDirName)
	if !changed {
		return false, nil
	}

	if err := document.Write(docPath, updated); err != nil {
		return false, fmt.Errorf("failed to update frontend profile dependencies in %s: %w", docPath, err)
	}

	return true, nil
}

func seedFrontendProfileDependencyRows(content string, docType document.DocumentType, featureDirName string) string {
	updated, _ := appendFrontendProfileDependencyRows(content, docType, featureDirName)
	return updated
}

func appendFrontendProfileDependencyRows(content string, docType document.DocumentType, featureDirName string) (string, bool) {
	return appendDependencyRowsToDocument(content, docType, canonicalFrontendProfileDependencyRows(featureDirName))
}

func hasActiveFrontendProfileDependency(section *document.Section) bool {
	for _, row := range dependencyRowsFromSection(section) {
		if dependencyCellMatches(row.Dependency, frontendProfileDependencyName) &&
			dependencyCellMatches(row.Type, frontendProfileDependencyType) &&
			dependencyCellMatches(row.Location, frontendProfileDependencyLocation) &&
			strings.EqualFold(normalizeDependencyCell(row.Status), "active") {
			return true
		}
	}
	return false
}

func canonicalFrontendProfileDependencyRows(featureDirName string) []profileDependencyRow {
	return []profileDependencyRow{
		{
			Dependency: frontendProfileDependencyName,
			Type:       frontendProfileDependencyType,
			Location:   frontendProfileDependencyLocation,
			UsedFor:    "apply frontend-specific coding-agent instruction set",
			Status:     "active",
		},
		{
			Dependency: designMaterialsDependencyName,
			Type:       designMaterialsDependencyType,
			Location:   designMaterialsRelPath(featureDirName),
			UsedFor:    "optional frontend design input",
			Status:     "optional",
		},
	}
}

func frontendProfileDependencySources(featurePath string) []struct {
	path    string
	docType document.DocumentType
} {
	return []struct {
		path    string
		docType document.DocumentType
	}{
		{path: filepath.Join(featurePath, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{path: filepath.Join(featurePath, "SPEC.md"), docType: document.TypeSpec},
		{path: filepath.Join(featurePath, "PLAN.md"), docType: document.TypePlan},
	}
}

func appendDependencyRowsToDocument(content string, docType document.DocumentType, rows []profileDependencyRow) (string, bool) {
	doc := document.Parse(content, "", docType)
	section := doc.GetSection("DEPENDENCIES")
	if section == nil {
		sectionBody := dependencyTableForRows(rows)
		trimmed := strings.TrimRight(content, "\n")
		return trimmed + "\n\n## DEPENDENCIES\n\n" + sectionBody + "\n", true
	}

	sectionBody, changed := appendDependencyRowsToSection(section.Content, rows)
	if !changed {
		return content, false
	}

	updated, ok := replaceMarkdownSection(content, "DEPENDENCIES", sectionBody)
	if !ok || updated == content {
		return content, false
	}
	return updated, true
}

func appendDependencyRowsToSection(sectionContent string, rows []profileDependencyRow) (string, bool) {
	lines := strings.Split(sectionContent, "\n")
	tableStart, tableEnd := dependencyTableBounds(lines)
	if tableStart == -1 {
		table := dependencyTableForRows(rows)
		trimmed := strings.TrimRight(sectionContent, "\n")
		if strings.TrimSpace(trimmed) == "" {
			return table, true
		}
		return trimmed + "\n\n" + table, true
	}

	tableLines := append([]string{}, lines[tableStart:tableEnd+1]...)
	tableLines, refreshed := refreshCanonicalDependencyRows(tableLines, rows)
	tableRows := dependencyTableRowsFromLines(tableLines)
	rowsToAdd := missingDependencyRows(tableRows, rows)
	shouldRemoveNoneRows := len(rowsToAdd) > 0 || dependencyTableHasAnyRow(tableRows, rows)
	placeholderRemoved := false

	if shouldRemoveNoneRows {
		before := len(tableLines)
		tableLines = removePlaceholderDependencyRows(tableLines)
		placeholderRemoved = len(tableLines) != before
	}
	if !refreshed && !placeholderRemoved && len(rowsToAdd) == 0 {
		return sectionContent, false
	}

	for _, row := range rowsToAdd {
		tableLines = append(tableLines, dependencyRowMarkdown(row))
	}

	updatedLines := make([]string, 0, len(lines)+len(rowsToAdd))
	updatedLines = append(updatedLines, lines[:tableStart]...)
	updatedLines = append(updatedLines, tableLines...)
	updatedLines = append(updatedLines, lines[tableEnd+1:]...)
	return strings.Join(updatedLines, "\n"), true
}

func dependencyTableForRows(rows []profileDependencyRow) string {
	lines := []string{
		"| Dependency | Type | Location | Used For | Status |",
		"| ---------- | ---- | -------- | -------- | ------ |",
	}
	for _, row := range rows {
		lines = append(lines, dependencyRowMarkdown(row))
	}
	return strings.Join(lines, "\n")
}

func refreshCanonicalDependencyRows(lines []string, rows []profileDependencyRow) ([]string, bool) {
	updated := append([]string{}, lines...)
	changed := false
	for i, line := range updated {
		parsedRows := dependencyTableRowsFromLines([]string{line})
		if len(parsedRows) != 1 {
			continue
		}
		for _, row := range rows {
			if !dependencyRowMatchesCanonicalIdentity(parsedRows[0], row) {
				continue
			}
			canonicalLine := dependencyRowMarkdown(row)
			if strings.TrimSpace(line) != canonicalLine {
				updated[i] = canonicalLine
				changed = true
			}
			break
		}
	}
	return updated, changed
}

func dependencyTableBounds(lines []string) (int, int) {
	start := -1
	end := -1
	for i, line := range lines {
		if !isDependencyTableLine(line) {
			continue
		}
		if start == -1 {
			start = i
		}
		end = i
	}
	return start, end
}

func isDependencyTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
}

func dependencyRowsFromSection(section *document.Section) []profileDependencyRow {
	if section == nil {
		return nil
	}
	rows := dependencyTableRows(section.Content)
	if len(rows) < 3 {
		return nil
	}

	header := dependencyHeaderIndex(rows[0])
	required := []string{"dependency", "type", "location", "used for", "status"}
	for _, key := range required {
		if _, ok := header[key]; !ok {
			return nil
		}
	}

	var result []profileDependencyRow
	for _, row := range rows[2:] {
		dependency := dependencyCell(row, header["dependency"])
		if dependency == "" || strings.EqualFold(normalizeDependencyCell(dependency), "none") {
			continue
		}
		result = append(result, profileDependencyRow{
			Dependency: dependency,
			Type:       dependencyCell(row, header["type"]),
			Location:   dependencyCell(row, header["location"]),
			UsedFor:    dependencyCell(row, header["used for"]),
			Status:     dependencyCell(row, header["status"]),
		})
	}
	return result
}

func dependencyTableRows(content string) [][]string {
	return dependencyTableRowsFromLines(strings.Split(content, "\n"))
}

func dependencyTableRowsFromLines(lines []string) [][]string {
	var rows [][]string
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "|") || !strings.Contains(strings.Trim(line, "|"), "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, cells)
	}
	return rows
}

func dependencyHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, cell := range header {
		index[strings.ToLower(strings.TrimSpace(cell))] = i
	}
	return index
}

func dependencyCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func missingDependencyRows(existing [][]string, rows []profileDependencyRow) []profileDependencyRow {
	var missing []profileDependencyRow
	for _, row := range rows {
		if !dependencyTableHasRow(existing, row) {
			missing = append(missing, row)
		}
	}
	return missing
}

func dependencyTableHasAnyRow(existing [][]string, rows []profileDependencyRow) bool {
	for _, row := range rows {
		if dependencyTableHasRow(existing, row) {
			return true
		}
	}
	return false
}

func dependencyTableHasRow(existing [][]string, row profileDependencyRow) bool {
	want := []string{row.Dependency, row.Type, row.Location, row.UsedFor, row.Status}
	for _, existingRow := range existing {
		if len(existingRow) < len(want) {
			continue
		}
		matches := true
		for i, cell := range want {
			if !dependencyCellMatches(existingRow[i], cell) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func dependencyRowMatchesCanonicalIdentity(existingRow []string, row profileDependencyRow) bool {
	if len(existingRow) < 5 {
		return false
	}
	return dependencyCellMatches(existingRow[0], row.Dependency) &&
		dependencyCellMatches(existingRow[1], row.Type) &&
		dependencyCellMatches(existingRow[2], row.Location) &&
		dependencyCellMatches(existingRow[4], row.Status)
}

func removePlaceholderDependencyRows(lines []string) []string {
	updated := make([]string, 0, len(lines))
	for _, line := range lines {
		rows := dependencyTableRowsFromLines([]string{line})
		if len(rows) == 1 && len(rows[0]) > 0 && strings.EqualFold(normalizeDependencyCell(rows[0][0]), "none") {
			continue
		}
		updated = append(updated, line)
	}
	return updated
}

func dependencyRowMarkdown(row profileDependencyRow) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s |", row.Dependency, row.Type, row.Location, row.UsedFor, row.Status)
}

func dependencyCellMatches(got, want string) bool {
	return normalizeDependencyCell(got) == normalizeDependencyCell(want)
}

func normalizeDependencyCell(value string) string {
	trimmed := strings.TrimSpace(value)
	for strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") && len(trimmed) >= 2 {
		trimmed = strings.TrimSpace(strings.Trim(trimmed, "`"))
	}
	return strings.ToLower(trimmed)
}
