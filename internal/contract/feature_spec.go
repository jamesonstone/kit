package contract

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
	"gopkg.in/yaml.v3"
)

var featureNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

var requiredFeatureSpecSections = []string{
	"purpose",
	"context",
	"source evidence and historical relationships",
	"requirements",
	"non-goals",
	"acceptance criteria",
	"accepted plan",
	"architecture and decisions",
	"discoveries",
	"validation plan and map",
	"validation",
	"outcome",
	"delivery evidence",
	"repository memory",
}

var sourcePhases = map[string]bool{
	"ready": true, "implement": true, "validate": true,
	"reflect": true, "deliver": true, "complete": true,
}

type specMetadata struct {
	KitMetadataVersion int                `yaml:"kit_metadata_version"`
	Artifact           string             `yaml:"artifact"`
	WorkflowVersion    int                `yaml:"workflow_version"`
	Phase              string             `yaml:"phase"`
	Feature            specFeature        `yaml:"feature"`
	Relationships      []specRelationship `yaml:"relationships"`
}

type specFeature struct {
	Dir string `yaml:"dir"`
}

type specRelationship struct {
	Target string `yaml:"target"`
}

func inspectFeatureSpec(root, feature string) (FeatureSpec, string) {
	required := append([]string(nil), requiredFeatureSpecSections...)
	result := FeatureSpec{
		Feature: feature, State: "missing", RequiredSections: required,
		MissingSections: append([]string(nil), required...),
		HistoricalSpecs: []string{}, HistoryIndexes: existingHistoryIndexes(root),
		PhasePermissions: PhasePermissions{SpecAuthoring: true},
	}
	if !validFeatureName(feature) {
		result.State = "invalid"
		return result, fmt.Sprintf("feature %q is not a safe canonical directory name", feature)
	}
	result.Path = path.Join("docs/specs", feature, "SPEC.md")
	content, exists, err := registry.ReadOptional(root, result.Path)
	if err != nil {
		result.State = "invalid"
		return result, fmt.Sprintf("feature spec %s cannot be read: %v", result.Path, err)
	}
	if !exists {
		return result, fmt.Sprintf("feature spec %s is missing", result.Path)
	}
	doc, err := registry.ParseMarkdown(content)
	if err != nil {
		result.State = "invalid"
		return result, fmt.Sprintf("feature spec %s is invalid: %v", result.Path, err)
	}
	metadata, err := parseSpecMetadata(content)
	if err != nil || metadata.KitMetadataVersion != 1 || metadata.Artifact != "spec" || metadata.WorkflowVersion != 3 ||
		metadata.Feature.Dir != feature || !validSpecPhase(metadata.Phase) {
		result.State = "invalid"
		return result, fmt.Sprintf("feature spec %s requires kit_metadata_version 1, artifact spec, workflow_version 3, a matching feature.dir, and a valid phase", result.Path)
	}
	result.WorkflowVersion = metadata.WorkflowVersion
	result.Phase = metadata.Phase
	result.HistoricalSpecs = relatedHistoricalSpecs(root, feature, metadata.Relationships)
	result.MissingSections = missingSpecSections(doc)
	if len(result.MissingSections) > 0 {
		result.State = "incomplete"
		return result, fmt.Sprintf("feature spec %s is missing required content: %s", result.Path, strings.Join(result.MissingSections, ", "))
	}
	result.State = "ready"
	result.PhasePermissions.SourceImplementation = sourcePhases[metadata.Phase]
	result.PhasePermissions.Delivery = metadata.Phase == "deliver" || metadata.Phase == "complete"
	if !result.PhasePermissions.SourceImplementation {
		return result, fmt.Sprintf("feature spec %s phase %q does not permit source implementation", result.Path, metadata.Phase)
	}
	return result, ""
}

func parseSpecMetadata(content string) (specMetadata, error) {
	if !strings.HasPrefix(content, "---\n") {
		return specMetadata{}, fmt.Errorf("front matter is required")
	}
	rest := strings.TrimPrefix(content, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return specMetadata{}, fmt.Errorf("front matter is not closed")
	}
	var metadata specMetadata
	if err := yaml.Unmarshal([]byte(rest[:end]), &metadata); err != nil {
		return specMetadata{}, err
	}
	metadata.Phase = strings.ToLower(strings.TrimSpace(metadata.Phase))
	return metadata, nil
}

func missingSpecSections(doc registry.MarkdownDocument) []string {
	missing := []string{}
	for _, required := range requiredFeatureSpecSections {
		section, exists := doc.Sections[required]
		lines := strings.SplitN(section, "\n", 2)
		if !exists || len(lines) < 2 || strings.TrimSpace(lines[1]) == "" {
			missing = append(missing, required)
		}
	}
	return missing
}

func relatedHistoricalSpecs(root, current string, relationships []specRelationship) []string {
	seen := map[string]bool{}
	var result []string
	for _, relationship := range relationships {
		candidate := historicalSpecPath(relationship.Target)
		if candidate == "" || candidate == path.Join("docs/specs", current, "SPEC.md") || seen[candidate] {
			continue
		}
		if _, exists, err := registry.ReadOptional(root, candidate); err == nil && exists {
			seen[candidate] = true
			result = append(result, candidate)
		}
	}
	sort.Strings(result)
	return result
}

func historicalSpecPath(target string) string {
	target = strings.TrimSpace(strings.TrimSuffix(target, "/"))
	if validFeatureName(target) {
		return path.Join("docs/specs", target, "SPEC.md")
	}
	if strings.HasPrefix(target, "docs/specs/") && strings.HasSuffix(target, "/SPEC.md") {
		middle := strings.TrimSuffix(strings.TrimPrefix(target, "docs/specs/"), "/SPEC.md")
		if validFeatureName(middle) {
			return target
		}
	}
	return ""
}

func existingHistoryIndexes(root string) []string {
	result := []string{"docs/specs/"}
	if _, exists, err := registry.ReadOptional(root, "docs/PROJECT_PROGRESS_SUMMARY.md"); err == nil && exists {
		result = append([]string{"docs/PROJECT_PROGRESS_SUMMARY.md"}, result...)
	}
	return result
}

func validFeatureName(feature string) bool {
	return featureNamePattern.MatchString(feature) && feature != "." && feature != ".."
}

func validSpecPhase(phase string) bool {
	return phase == "clarify" || phase == "blocked" || sourcePhases[phase]
}
