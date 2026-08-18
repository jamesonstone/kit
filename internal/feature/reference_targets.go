package feature

import (
	"net/url"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/document"
)

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
