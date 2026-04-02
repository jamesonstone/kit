package document

import "strings"

var explicitSectionFallbackTexts = map[string]struct{}{
	"not applicable":                     {},
	"not required":                       {},
	"no additional information required": {},
}

func documentTypeRequiresPopulatedSections(docType DocumentType) bool {
	switch docType {
	case TypeBrainstorm, TypeSpec, TypePlan, TypeTasks:
		return true
	default:
		return false
	}
}

func sectionHasVisibleContent(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if visibleLineContent(line) != "" {
			return true
		}
	}

	return false
}

func visibleLineContent(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "<!--") {
		if idx := strings.Index(trimmed, "-->"); idx != -1 {
			return strings.TrimSpace(trimmed[idx+3:])
		}
		return ""
	}

	if idx := strings.Index(trimmed, "<!--"); idx != -1 {
		trimmed = strings.TrimSpace(trimmed[:idx])
	}

	return trimmed
}

func isExplicitSectionFallbackText(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	normalized = strings.TrimSuffix(normalized, ".")

	_, ok := explicitSectionFallbackTexts[normalized]
	return ok
}
