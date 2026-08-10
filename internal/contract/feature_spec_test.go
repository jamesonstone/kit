package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestResolveFeatureSpecBlocksSourceButPermitsAuthoring(t *testing.T) {
	root := writeContractProject(t)
	hints := Hints{WorkType: WorkTypeFeature, Feature: "0059-example", Workflows: []string{"delivery"}}

	missing := resolveFeature(t, root, hints)
	assertFeatureBlocked(t, missing, "missing")
	if missing.FeatureSpec.Path != "docs/specs/0059-example/SPEC.md" {
		t.Fatalf("feature spec path = %q", missing.FeatureSpec.Path)
	}

	writeContractFile(t, filepath.Join(root, missing.FeatureSpec.Path), featureSpecDocument("0059-example", "ready", []string{"purpose"}, nil))
	incomplete := resolveFeature(t, root, hints)
	assertFeatureBlocked(t, incomplete, "incomplete")
	if !slices.Contains(incomplete.FeatureSpec.MissingSections, "repository memory") {
		t.Fatalf("missing sections = %v", incomplete.FeatureSpec.MissingSections)
	}

	invalid := strings.Replace(featureSpecDocument("0059-example", "ready", requiredFeatureSpecSections, nil), "workflow_version: 3", "workflow_version: 2", 1)
	writeContractFile(t, filepath.Join(root, missing.FeatureSpec.Path), invalid)
	assertFeatureBlocked(t, resolveFeature(t, root, hints), "invalid")
}

func TestResolveCompleteFeatureSpecUnlocksSourceAndDiscoversHistory(t *testing.T) {
	root := writeContractProject(t)
	writeContractFile(t, filepath.Join(root, "docs/PROJECT_PROGRESS_SUMMARY.md"), "# Progress\n")
	writeContractFile(t, filepath.Join(root, "docs/specs/0042-history/SPEC.md"), "# Historical\n")
	path := filepath.Join(root, "docs/specs/0059-example/SPEC.md")
	writeContractFile(t, path, featureSpecDocument("0059-example", "ready", requiredFeatureSpecSections, []string{"0042-history"}))

	resolved := resolveFeature(t, root, Hints{WorkType: WorkTypeFeature, Feature: "0059-example", Workflows: []string{"delivery"}})
	if resolved.State != "ready" || resolved.FeatureSpec.State != "ready" {
		t.Fatalf("resolved state = %s, feature spec = %#v", resolved.State, resolved.FeatureSpec)
	}
	permissions := resolved.FeatureSpec.PhasePermissions
	if !permissions.SpecAuthoring || !permissions.SourceImplementation || permissions.Delivery {
		t.Fatalf("phase permissions = %#v", permissions)
	}
	if !slices.Contains(resolved.FeatureSpec.HistoricalSpecs, "docs/specs/0042-history/SPEC.md") ||
		!slices.Contains(resolved.FeatureSpec.HistoryIndexes, "docs/PROJECT_PROGRESS_SUMMARY.md") {
		t.Fatalf("history = %v, indexes = %v", resolved.FeatureSpec.HistoricalSpecs, resolved.FeatureSpec.HistoryIndexes)
	}
	if !hasArtifact(resolved.Rules["conditional"], "feature-specification") {
		t.Fatalf("feature-specification dependency not selected: %#v", resolved.Rules)
	}
	assertFeatureSpecGolden(t, resolved.FeatureSpec)

	writeContractFile(t, path, featureSpecDocument("0059-example", "deliver", requiredFeatureSpecSections, nil))
	deliver := resolveFeature(t, root, Hints{WorkType: WorkTypeFeature, Feature: "0059-example", Workflows: []string{"delivery"}})
	if !deliver.FeatureSpec.PhasePermissions.Delivery {
		t.Fatalf("delivery permission = %#v", deliver.FeatureSpec.PhasePermissions)
	}
}

func assertFeatureSpecGolden(t *testing.T, featureSpec *FeatureSpec) {
	t.Helper()
	got, err := json.MarshalIndent(featureSpec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/feature-spec.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(append(got, '\n')) != string(want) {
		t.Fatalf("feature spec differs from golden\n--- got\n%s\n--- want\n%s", got, want)
	}
}

func TestFeatureSpecV3RequiresDetailedLifecycleSections(t *testing.T) {
	want := []string{
		"purpose", "context", "source evidence and historical relationships",
		"requirements", "non-goals", "acceptance criteria", "accepted plan",
		"architecture and decisions", "discoveries", "validation plan and map",
		"validation", "outcome", "delivery evidence", "repository memory",
	}
	if !slices.Equal(requiredFeatureSpecSections, want) {
		t.Fatalf("required sections = %v", requiredFeatureSpecSections)
	}
}

func TestRepositoryFeatureSpecExemplarIsStructurallyComplete(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	resolved := resolveFeature(t, root, Hints{
		WorkType: WorkTypeFeature, Feature: "0058-coding-agent-first", Workflows: []string{"implementation-delivery"},
	})
	if resolved.State != "ready" || resolved.FeatureSpec.State != "ready" ||
		len(resolved.FeatureSpec.MissingSections) != 0 || !resolved.FeatureSpec.PhasePermissions.SourceImplementation {
		t.Fatalf("repository feature spec = %#v", resolved.FeatureSpec)
	}
	if len(resolved.FeatureSpec.HistoricalSpecs) == 0 {
		t.Fatal("repository feature spec has no discoverable historical relationships")
	}
}

func TestResolveRejectsUnsafeFeatureHint(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{WorkType: WorkTypeFeature, Feature: "../escape", Workflows: []string{"delivery"}})
	assertFeatureBlocked(t, resolved, "invalid")
	if resolved.FeatureSpec.Path != "" {
		t.Fatalf("unsafe feature path = %q", resolved.FeatureSpec.Path)
	}
}

func resolveFeature(t *testing.T, root string, hints Hints) Resolved {
	t.Helper()
	resolved, err := Resolve(root, hints)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func assertFeatureBlocked(t *testing.T, resolved Resolved, state string) {
	t.Helper()
	if resolved.State != "blocked" || resolved.FeatureSpec == nil || resolved.FeatureSpec.State != state {
		t.Fatalf("resolved state = %s, feature spec = %#v", resolved.State, resolved.FeatureSpec)
	}
	permissions := resolved.FeatureSpec.PhasePermissions
	if !permissions.SpecAuthoring || permissions.SourceImplementation || permissions.Delivery {
		t.Fatalf("phase permissions = %#v", permissions)
	}
}

func featureSpecDocument(feature, phase string, sections, relationships []string) string {
	content := "---\nkit_metadata_version: 1\nartifact: spec\nworkflow_version: 3\nphase: " + phase + "\nfeature:\n  dir: " + feature + "\n"
	if len(relationships) > 0 {
		content += "relationships:\n"
		for _, target := range relationships {
			content += "  - type: builds_on\n    target: " + target + "\n"
		}
	}
	content += "---\n# SPEC\n"
	for _, section := range sections {
		content += fmt.Sprintf("\n## %s\n\nVerified %s content.\n", strings.ToUpper(section), section)
	}
	return content
}
