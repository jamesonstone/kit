package contract

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
)

func Resolve(root string, hints Hints) (Resolved, error) {
	cfg, migration, err := registry.LoadProject(root)
	if err != nil {
		return Resolved{}, err
	}
	resolved := Resolved{
		SchemaVersion: registry.ContractSchemaVersion,
		State:         "ready",
		ProjectRoot:   root,
		Hints:         normalizedHints(hints),
		Registry:      cfg.Registry.Source,
		Routing:       []string{},
		Workflows:     []Artifact{},
		Rules:         map[string][]Artifact{},
	}
	if migration {
		resolved.State = "blocked"
		resolved.Diagnostics = append(resolved.Diagnostics, ".kit.yaml requires schema-v2 migration")
		resolved.NextActions = append(resolved.NextActions, "run `kit reconcile` to preview the migration, then apply it explicitly")
		return resolved, nil
	}
	selected, reasons := selectRecords(cfg.Registry.Artifacts, resolved.Hints)
	selected = includeDependencies(cfg.Registry.Artifacts, selected, reasons, &resolved)
	for _, record := range selected {
		key := registry.ArtifactKey(record.Kind, record.Slug)
		artifact, blocking := resolveRecord(root, record, reasons[key])
		if record.Kind == registry.KindWorkflow {
			resolved.Workflows = append(resolved.Workflows, artifact)
		} else {
			policy := record.ReadPolicy
			if policy == "" {
				policy = "conditional"
			}
			resolved.Rules[policy] = append(resolved.Rules[policy], artifact)
		}
		if blocking {
			resolved.State = "blocked"
			resolved.Diagnostics = append(resolved.Diagnostics, fmt.Sprintf("%s/%s is %s", record.Kind, record.Slug, artifact.State))
		}
	}
	for _, workflow := range resolved.Hints.Workflows {
		if !hasArtifact(resolved.Workflows, workflow) {
			resolved.State = "blocked"
			resolved.Diagnostics = append(resolved.Diagnostics, fmt.Sprintf("requested workflow %q is not installed", workflow))
		}
	}
	resolved.Routing = existingRouting(root)
	sortResolved(&resolved)
	if resolved.State == "blocked" {
		resolved.NextActions = appendUnique(resolved.NextActions, "run `kit reconcile` and resolve required artifact diagnostics")
	} else {
		resolved.NextActions = append(resolved.NextActions, "read the selected workflows and ordered rules before implementation")
	}
	return resolved, nil
}

func selectRecords(records []registry.ArtifactRecord, hints Hints) ([]registry.ArtifactRecord, map[string]string) {
	var selected []registry.ArtifactRecord
	reasons := map[string]string{}
	for _, record := range records {
		reason := selectionReason(record, hints)
		if reason != "" {
			selected = append(selected, record)
			reasons[registry.ArtifactKey(record.Kind, record.Slug)] = reason
		}
	}
	return selected, reasons
}

func selectionReason(record registry.ArtifactRecord, hints Hints) string {
	if record.Kind == registry.KindWorkflow && contains(hints.Workflows, record.Slug) {
		return "explicit workflow"
	}
	if record.ReadPolicy == "must" {
		return "mandatory"
	}
	for _, tag := range hints.Applicability {
		if contains(record.AppliesTo, tag) {
			return "applicability tag " + tag
		}
	}
	if hints.Feature != "" && contains(record.AppliesTo, hints.Feature) {
		return "feature " + hints.Feature
	}
	for _, candidate := range hints.Paths {
		for _, pattern := range record.Paths {
			if matchPath(pattern, candidate) {
				return "path " + candidate
			}
		}
	}
	return ""
}

func includeDependencies(all, selected []registry.ArtifactRecord, reasons map[string]string, resolved *Resolved) []registry.ArtifactRecord {
	byKey := map[string]registry.ArtifactRecord{}
	chosen := map[string]bool{}
	for _, record := range all {
		byKey[registry.ArtifactKey(record.Kind, record.Slug)] = record
	}
	for _, record := range selected {
		chosen[registry.ArtifactKey(record.Kind, record.Slug)] = true
	}
	for changed := true; changed; {
		changed = false
		for _, record := range append([]registry.ArtifactRecord(nil), selected...) {
			for _, dependency := range record.Dependencies {
				key := dependency
				if !strings.Contains(key, "/") {
					key = registry.ArtifactKey(registry.KindRuleset, key)
				}
				if chosen[key] {
					continue
				}
				dependencyRecord, ok := byKey[key]
				if !ok {
					resolved.State = "blocked"
					resolved.Diagnostics = append(resolved.Diagnostics, fmt.Sprintf("%s depends on missing %s", registry.ArtifactKey(record.Kind, record.Slug), key))
					continue
				}
				selected = append(selected, dependencyRecord)
				chosen[key] = true
				reasons[key] = "dependency of " + registry.ArtifactKey(record.Kind, record.Slug)
				changed = true
			}
		}
	}
	return selected
}

func resolveRecord(root string, record registry.ArtifactRecord, reason string) (Artifact, bool) {
	artifact := Artifact{
		Kind: record.Kind, Slug: record.Slug, Description: record.Description,
		Path: record.Path, Version: record.Version, ReadPolicy: record.ReadPolicy,
		State: record.State, SourceRepo: record.SourceRepo, SourceBranch: record.SourceBranch,
		SourceCommit: record.SourceCommit, SourcePath: record.SourcePath,
		Dependencies:  append([]string(nil), record.Dependencies...),
		Applicability: append([]string(nil), record.AppliesTo...), Reason: reason,
	}
	content, exists, err := registry.ReadOptional(root, record.Path)
	if err != nil || !exists {
		artifact.State = registry.StateMissing
		return artifact, true
	}
	artifact.Digest = registry.HashContent(content)
	doc, parseErr := registry.ParseMarkdown(content)
	if parseErr != nil || doc.Metadata.Kind != record.Kind || doc.Metadata.Slug != record.Slug {
		artifact.State = "invalid"
		return artifact, true
	}
	if artifact.State == registry.StateManaged && record.ContentHash != "" && artifact.Digest != record.ContentHash {
		artifact.State = registry.StateLocalCustom
	}
	blocking := artifact.State == registry.StateConflict || artifact.State == registry.StateMissing || artifact.State == "invalid"
	return artifact, blocking
}

func matchPath(pattern, candidate string) bool {
	pattern = strings.TrimPrefix(path.Clean("/"+pattern), "/")
	candidate = strings.TrimPrefix(path.Clean("/"+candidate), "/")
	if matched, _ := path.Match(pattern, candidate); matched {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return candidate == prefix || strings.HasPrefix(candidate, prefix+"/")
	}
	if !strings.HasPrefix(pattern, "**/") {
		return false
	}
	tail := strings.TrimPrefix(pattern, "**/")
	for current := candidate; ; {
		if matched, _ := path.Match(tail, current); matched {
			return true
		}
		separator := strings.Index(current, "/")
		if separator < 0 {
			return false
		}
		current = current[separator+1:]
	}
}

func existingRouting(root string) []string {
	paths := []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md", "docs/agents/README.md"}
	result := []string{}
	for _, candidate := range paths {
		if _, exists, _ := registry.ReadOptional(root, candidate); exists {
			result = append(result, candidate)
		}
	}
	return result
}

func sortResolved(resolved *Resolved) {
	sort.Slice(resolved.Workflows, func(i, j int) bool { return resolved.Workflows[i].Slug < resolved.Workflows[j].Slug })
	for policy := range resolved.Rules {
		sort.Slice(resolved.Rules[policy], func(i, j int) bool { return resolved.Rules[policy][i].Slug < resolved.Rules[policy][j].Slug })
	}
	sort.Strings(resolved.Diagnostics)
}

func normalizedHints(hints Hints) Hints {
	hints.Feature = strings.TrimSpace(hints.Feature)
	hints.Paths = uniqueSorted(hints.Paths)
	hints.Applicability = uniqueSorted(hints.Applicability)
	hints.Workflows = uniqueSorted(hints.Workflows)
	return hints
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func hasArtifact(artifacts []Artifact, slug string) bool {
	for _, artifact := range artifacts {
		if artifact.Slug == slug {
			return true
		}
	}
	return false
}

func appendUnique(values []string, value string) []string {
	if contains(values, value) {
		return values
	}
	return append(values, value)
}
