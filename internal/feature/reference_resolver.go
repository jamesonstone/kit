package feature

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

type referenceResolution struct {
	NodeID   string
	Kind     string
	Resolved bool
	Error    string
}

func resolveReference(projectRoot string, cfg *config.Config, reference document.MetadataReference) referenceResolution {
	target := cleanReferenceTarget(reference.Target)
	selectorType := strings.TrimSpace(reference.SelectorType)
	selector := strings.TrimSpace(reference.Selector)
	nodeID := referenceNodeID(reference)

	if target == "" {
		return unresolvedReference(nodeID, "target is empty")
	}
	if isExternalReferenceTarget(target) {
		return referenceResolution{NodeID: nodeID, Kind: "external_url", Resolved: true}
	}
	if strings.EqualFold(target, "n/a") {
		return unresolvedReference(nodeID, "target is not applicable")
	}
	if isLogicalReference(reference, target) {
		return resolveLogicalReference(nodeID, reference, target)
	}
	if components := referenceTargetComponents(target); len(components) > 1 {
		return resolveCompositeReference(projectRoot, cfg, nodeID, components)
	}

	resolvedPath, exists, isDir := resolveReferencePath(projectRoot, cfg, target)
	if !exists {
		return unresolvedReference(nodeID, fmt.Sprintf("target %q does not exist", target))
	}

	if selector == "" {
		if isDir {
			return referenceResolution{NodeID: nodeID, Kind: "directory", Resolved: true}
		}
		return referenceResolution{NodeID: nodeID, Kind: "file", Resolved: true}
	}

	switch selectorType {
	case document.ReferenceSelectorTypeArtifact:
		return resolveArtifactSelector(nodeID, resolvedPath, isDir, selector)
	case document.ReferenceSelectorTypeHeading:
		return resolveContentSelector(nodeID, resolvedPath, isDir, selector, "heading", fileContainsHeading)
	case document.ReferenceSelectorTypeSymbol:
		return resolveContentSelector(nodeID, resolvedPath, isDir, selector, "symbol", fileContainsSymbol)
	case document.ReferenceSelectorTypeCommand:
		return resolveContentSelector(nodeID, resolvedPath, isDir, selector, "command", fileContainsCommand)
	case document.ReferenceSelectorTypeURL, document.ReferenceSelectorTypeNodeID:
		return referenceResolution{NodeID: nodeID, Kind: selectorType, Resolved: true}
	case "":
		return unresolvedReference(nodeID, "selector is set without selector_type")
	default:
		return unresolvedReference(nodeID, fmt.Sprintf("unsupported selector_type %q", selectorType))
	}
}

func resolveReferencePath(projectRoot string, cfg *config.Config, target string) (string, bool, bool) {
	target = normalizeReferenceTargetComponent(target)
	candidates := []string{}
	if filepath.IsAbs(target) {
		candidates = append(candidates, filepath.Clean(target))
	} else if strings.HasPrefix(target, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(target, "~/"))))
		}
	} else {
		candidates = append(candidates, filepath.Join(projectRoot, filepath.FromSlash(target)))
		if !strings.Contains(target, "/") {
			candidates = append(candidates, filepath.Join(cfg.SpecsPath(projectRoot), target))
		}
	}

	if hasReferenceGlob(target) {
		for _, candidate := range candidates {
			matches, err := filepath.Glob(candidate)
			if err == nil && len(matches) > 0 {
				return candidate, true, false
			}
		}
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil {
			return candidate, true, info.IsDir()
		}
	}

	if len(candidates) == 0 {
		return "", false, false
	}
	return candidates[0], false, false
}

func resolveCompositeReference(projectRoot string, cfg *config.Config, nodeID string, components []string) referenceResolution {
	var unresolved []string
	for _, component := range components {
		_, exists, _ := resolveReferencePath(projectRoot, cfg, component)
		if exists {
			continue
		}
		unresolved = append(unresolved, normalizeReferenceTargetComponent(component))
	}
	if len(unresolved) == 0 {
		return referenceResolution{NodeID: nodeID, Kind: "composite", Resolved: true}
	}
	return unresolvedReference(nodeID, fmt.Sprintf("unresolved target component(s): %s", strings.Join(unresolved, ", ")))
}

func isLogicalReference(reference document.MetadataReference, target string) bool {
	referenceType := strings.ToLower(strings.TrimSpace(reference.Type))
	if strings.Contains(referenceType, "command") ||
		strings.Contains(referenceType, "tool") ||
		strings.Contains(referenceType, "profile") ||
		strings.Contains(referenceType, "library") ||
		strings.Contains(referenceType, "package") {
		return true
	}
	if strings.HasPrefix(target, "--") {
		return true
	}
	return isGoModuleReference(target)
}

func resolveLogicalReference(nodeID string, reference document.MetadataReference, target string) referenceResolution {
	referenceType := strings.ToLower(strings.TrimSpace(reference.Type))
	if strings.Contains(referenceType, "library") || strings.Contains(referenceType, "package") || isGoModuleReference(target) {
		return referenceResolution{NodeID: nodeID, Kind: "external_package", Resolved: true}
	}
	if strings.Contains(referenceType, "profile") || strings.HasPrefix(target, "--") {
		return referenceResolution{NodeID: nodeID, Kind: "profile", Resolved: true}
	}
	if strings.Contains(referenceType, "command") || strings.Contains(referenceType, "tool") {
		command := firstCommandWord(target)
		if command == "" {
			return unresolvedReference(nodeID, "command target is empty")
		}
		if command == "kit" {
			return referenceResolution{NodeID: nodeID, Kind: "command", Resolved: true}
		}
		if _, err := exec.LookPath(command); err == nil {
			return referenceResolution{NodeID: nodeID, Kind: "command", Resolved: true}
		}
		return unresolvedReference(nodeID, fmt.Sprintf("command %q not found on PATH", command))
	}
	return referenceResolution{NodeID: nodeID, Kind: "logical", Resolved: true}
}

func resolveArtifactSelector(nodeID, resolvedPath string, isDir bool, selector string) referenceResolution {
	if isDir {
		path := filepath.Join(resolvedPath, filepath.FromSlash(selector))
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return referenceResolution{NodeID: nodeID, Kind: "artifact", Resolved: true}
		}
		return unresolvedReference(nodeID, fmt.Sprintf("artifact %q does not exist", selector))
	}
	if filepath.Base(resolvedPath) == selector {
		return referenceResolution{NodeID: nodeID, Kind: "artifact", Resolved: true}
	}
	return unresolvedReference(nodeID, fmt.Sprintf("artifact selector %q does not match %q", selector, filepath.Base(resolvedPath)))
}

func resolveContentSelector(
	nodeID string,
	resolvedPath string,
	isDir bool,
	selector string,
	kind string,
	matcher func(string, string) bool,
) referenceResolution {
	if isDir {
		return unresolvedReference(nodeID, fmt.Sprintf("%s selector requires a file target", kind))
	}
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return unresolvedReference(nodeID, fmt.Sprintf("failed to read target: %v", err))
	}
	if matcher(string(content), selector) {
		return referenceResolution{NodeID: nodeID, Kind: kind, Resolved: true}
	}
	return unresolvedReference(nodeID, fmt.Sprintf("%s selector %q not found", kind, selector))
}

func fileContainsHeading(content string, selector string) bool {
	want := normalizeReferenceSelector(selector)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		if normalizeReferenceSelector(heading) == want {
			return true
		}
	}
	return false
}

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
