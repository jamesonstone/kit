package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

const featureNotesDependencyName = "Feature notes"

func featureNotesRelPath(featureDirName string) string {
	return filepath.ToSlash(filepath.Join("docs", "notes", featureDirName))
}

func featureNotesPath(projectRoot, featureDirName string) string {
	return filepath.Join(projectRoot, "docs", "notes", featureDirName)
}

func ensureFeatureNotesDir(projectRoot, featureDirName string) (string, string, error) {
	notesPath := featureNotesPath(projectRoot, featureDirName)
	if err := os.MkdirAll(notesPath, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create feature notes directory: %w", err)
	}

	gitkeepPath := filepath.Join(notesPath, ".gitkeep")
	file, err := os.OpenFile(gitkeepPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", "", fmt.Errorf("failed to create feature notes placeholder: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close feature notes placeholder: %w", err)
	}

	return notesPath, featureNotesRelPath(featureDirName), nil
}

func featureNotesDirName(brainstormPath, fallbackSlug string) string {
	dirName := filepath.Base(filepath.Dir(brainstormPath))
	if dirName == "" || dirName == "." || dirName == string(filepath.Separator) {
		return fallbackSlug
	}
	return dirName
}

func seedBrainstormNotesDependency(content, notesRelPath string) string {
	row := featureNotesDependencyRow(notesRelPath)
	defaultRow := "| none | n/a | n/a | no phase dependencies recorded yet | active |"
	if brainstormNotesDependencyExists(content, notesRelPath) {
		return content
	}
	if strings.Contains(content, defaultRow) {
		return strings.Replace(content, defaultRow, row, 1)
	}
	return content
}

func ensureBrainstormNotesDependency(brainstormPath, notesRelPath string) (bool, error) {
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", brainstormPath, err)
	}

	updated, changed := appendBrainstormNotesDependency(string(content), notesRelPath)
	if !changed {
		return false, nil
	}

	if err := document.Write(brainstormPath, updated); err != nil {
		return false, fmt.Errorf("failed to update notes dependency in %s: %w", brainstormPath, err)
	}

	return true, nil
}

func appendBrainstormNotesDependency(content, notesRelPath string) (string, bool) {
	if brainstormNotesDependencyExists(content, notesRelPath) {
		return content, false
	}

	row := featureNotesDependencyRow(notesRelPath)
	doc := document.Parse(content, "", document.TypeBrainstorm)
	section := doc.GetSection("DEPENDENCIES")
	if section == nil {
		sectionBody := strings.Join([]string{
			"| Dependency | Type | Location | Used For | Status |",
			"| ---------- | ---- | -------- | -------- | ------ |",
			row,
		}, "\n")
		trimmed := strings.TrimRight(content, "\n")
		return trimmed + "\n\n## DEPENDENCIES\n\n" + sectionBody + "\n", true
	}

	sectionBody, changed := appendDependencyTableRow(section.Content, row)
	if !changed {
		return content, false
	}

	updated, ok := replaceMarkdownSection(content, "DEPENDENCIES", sectionBody)
	if !ok || updated == content {
		return content, false
	}
	return updated, true
}

func appendDependencyTableRow(sectionContent, row string) (string, bool) {
	lines := strings.Split(sectionContent, "\n")
	lastTableLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			lastTableLine = i
		}
	}

	if lastTableLine == -1 {
		table := strings.Join([]string{
			"| Dependency | Type | Location | Used For | Status |",
			"| ---------- | ---- | -------- | -------- | ------ |",
			row,
		}, "\n")
		trimmed := strings.TrimRight(sectionContent, "\n")
		if strings.TrimSpace(trimmed) == "" {
			return table, true
		}
		return trimmed + "\n\n" + table, true
	}

	updatedLines := make([]string, 0, len(lines)+1)
	updatedLines = append(updatedLines, lines[:lastTableLine+1]...)
	updatedLines = append(updatedLines, row)
	updatedLines = append(updatedLines, lines[lastTableLine+1:]...)
	return strings.Join(updatedLines, "\n"), true
}

func featureNotesDependencyRow(notesRelPath string) string {
	return fmt.Sprintf(
		"| %s | notes | %s | optional pre-brainstorm research input | optional |",
		featureNotesDependencyName,
		notesRelPath,
	)
}

func brainstormNotesDependencyExists(content, notesRelPath string) bool {
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 3 {
			continue
		}
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if cells[0] == featureNotesDependencyName && cells[2] == notesRelPath {
			return true
		}
	}
	return false
}
