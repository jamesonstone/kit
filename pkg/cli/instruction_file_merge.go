package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var instructionFileSectionPattern = regexp.MustCompile(`(?m)^##\s+(.+)$`)

type parsedInstructionFile struct {
	preamble string
	sections []instructionFileSection
}

type instructionFileSection struct {
	name string
	key  string
	raw  string
}

func readInstructionFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func mergeInstructionFileContent(existingContent, templateContent string) (string, bool, error) {
	existing := parseInstructionFileContent(existingContent)
	template := parseInstructionFileContent(templateContent)

	if len(template.sections) == 0 {
		return "", false, fmt.Errorf("template has no top-level '##' sections")
	}

	templateIndex := make(map[string]int, len(template.sections))
	for i, section := range template.sections {
		templateIndex[section.key] = i
	}

	existingByKey := make(map[string]instructionFileSection, len(existing.sections))
	recognizedCount := 0
	for _, section := range existing.sections {
		if _, ok := templateIndex[section.key]; !ok {
			continue
		}
		recognizedCount++
		if _, exists := existingByKey[section.key]; exists {
			return "", false, fmt.Errorf("duplicate recognized section %q", section.name)
		}
		existingByKey[section.key] = section
	}

	if recognizedCount == 0 {
		return "", false, fmt.Errorf("no recognizable Kit-managed sections found")
	}

	missingCount := 0
	for _, section := range template.sections {
		if _, ok := existingByKey[section.key]; !ok {
			missingCount++
		}
	}

	if missingCount == 0 {
		return existingContent, false, nil
	}

	var builder strings.Builder
	preamble := existing.preamble
	if strings.TrimSpace(preamble) == "" {
		preamble = template.preamble
	}
	builder.WriteString(preamble)

	cursor := 0
	appendTemplateSectionsUntil := func(limit int) {
		for cursor < limit {
			section := template.sections[cursor]
			if _, ok := existingByKey[section.key]; !ok {
				builder.WriteString(section.raw)
			}
			cursor++
		}
	}

	for _, section := range existing.sections {
		index, ok := templateIndex[section.key]
		if !ok {
			builder.WriteString(section.raw)
			continue
		}

		appendTemplateSectionsUntil(index)
		builder.WriteString(section.raw)
		cursor = index + 1
	}

	appendTemplateSectionsUntil(len(template.sections))

	return builder.String(), true, nil
}

func parseInstructionFileContent(content string) parsedInstructionFile {
	matches := instructionFileSectionPattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return parsedInstructionFile{preamble: content}
	}

	parsed := parsedInstructionFile{
		preamble: content[:matches[0][0]],
		sections: make([]instructionFileSection, 0, len(matches)),
	}

	for i, match := range matches {
		start := match[0]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		name := strings.TrimSpace(content[match[2]:match[3]])
		parsed.sections = append(parsed.sections, instructionFileSection{
			name: name,
			key:  normalizeInstructionSectionName(name),
			raw:  content[start:end],
		})
	}

	return parsed
}

func normalizeInstructionSectionName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}
