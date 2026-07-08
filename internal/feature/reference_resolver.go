package feature

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
