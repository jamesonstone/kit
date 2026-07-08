package feature

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

func fileContainsSymbol(content string, selector string) bool {
	quoted := regexp.QuoteMeta(selector)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\s*func\s+(?:\([^)]+\)\s*)?` + quoted + `\b`),
		regexp.MustCompile(`(?m)^\s*type\s+` + quoted + `\b`),
		regexp.MustCompile(`(?m)^\s*(?:const|var)\s+(?:\([^)]*\b` + quoted + `\b|` + quoted + `\b)`),
	}
	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return strings.Contains(content, selector)
}

func fileContainsCommand(content string, selector string) bool {
	if strings.Contains(content, selector) {
		return true
	}
	parts := strings.Fields(selector)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts[1:] {
		if strings.Contains(content, part) {
			continue
		}
		if strings.HasPrefix(part, "--") && strings.Contains(content, strings.TrimPrefix(part, "--")) {
			continue
		}
		if strings.HasPrefix(part, "-") && strings.Contains(content, strings.TrimPrefix(part, "-")) {
			continue
		}
		return false
	}
	return len(parts) > 1
}

func cleanReferenceTarget(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "`\"'")
	return strings.TrimSpace(value)
}

func referenceTargetComponents(target string) []string {
	parts := strings.Split(target, ",")
	if len(parts) == 1 {
		component := normalizeReferenceTargetComponent(parts[0])
		if component == "" {
			return nil
		}
		return []string{component}
	}

	components := make([]string, 0, len(parts))
	for _, part := range parts {
		component := normalizeReferenceTargetComponent(part)
		if component == "" {
			continue
		}
		components = append(components, component)
	}
	return components
}

func normalizeReferenceTargetComponent(component string) string {
	component = cleanReferenceTarget(component)
	if component == "" {
		return ""
	}
	fields := strings.Fields(component)
	if len(fields) > 1 && looksLikePathishTarget(fields[0]) {
		return cleanReferenceTarget(fields[0])
	}
	return component
}

func looksLikePathishTarget(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" &&
		(strings.Contains(value, "/") ||
			strings.Contains(value, "\\") ||
			strings.HasPrefix(value, ".") ||
			strings.HasPrefix(value, "~") ||
			hasReferenceGlob(value))
}

func hasReferenceGlob(value string) bool {
	return strings.ContainsAny(value, "*?[")
}

func isGoModuleReference(target string) bool {
	if strings.ContainsAny(target, " `,") || strings.HasPrefix(target, ".") || strings.HasPrefix(target, "/") {
		return false
	}
	parts := strings.Split(target, "/")
	return len(parts) > 1 && strings.Contains(parts[0], ".")
}

func firstCommandWord(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	return strings.Fields(target)[0]
}

func isExternalReferenceTarget(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func referenceNodeID(reference document.MetadataReference) string {
	if strings.TrimSpace(reference.ID) != "" {
		return strings.TrimSpace(reference.ID)
	}
	parts := []string{
		strings.TrimSpace(reference.Target),
		strings.TrimSpace(reference.SelectorType),
		strings.TrimSpace(reference.Selector),
	}
	return strings.Join(parts, "#")
}

func normalizeReferenceSelector(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func unresolvedReference(nodeID string, message string) referenceResolution {
	return referenceResolution{NodeID: nodeID, Kind: "unresolved", Resolved: false, Error: message}
}
