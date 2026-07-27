package feature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestBuildProjectMap_ReadsFrontMatterRelationshipsAndReferences(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	for _, dirName := range []string{"0001-ui", "0002-api"} {
		if err := os.MkdirAll(filepath.Join(specsDir, dirName), 0755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dirName, err)
		}
	}

	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-ui", "SPEC.md"), `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: ui
  dir: 0001-ui
relationships:
  - type: depends_on
    target: 0002-api
references:
  - name: Design brief
    type: doc
    target: docs/notes/0001-ui/design/brief.md
    selector_type: heading
    selector: Constraints
    relation: constrains
    read_policy: must
    used_for: UI constraints
    status: active
---
# SPEC

## RELATIONSHIPS

Relationships are tracked in front matter.

## DEPENDENCIES

Dependencies are tracked in front matter.
`)
	writeMapFile(t, filepath.Join(projectRoot, "docs", "notes", "0001-ui", "design", "brief.md"), "# Brief\n\n## Constraints\n\nUse the existing design system.\n")
	writeMapFile(t, filepath.Join(specsDir, "0002-api", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}
	ui := featureMapByDirName(t, projectMap.Features, "0001-ui")
	if len(ui.Outgoing) != 1 || ui.Outgoing[0].Type != "depends on" || ui.Outgoing[0].TargetFeatureID != "0002-api" {
		t.Fatalf("unexpected outgoing relationships: %#v", ui.Outgoing)
	}
	if len(ui.References) != 1 || ui.References[0].Reference != "Design brief" || ui.References[0].Target != "docs/notes/0001-ui/design/brief.md" {
		t.Fatalf("unexpected references: %#v", ui.References)
	}
	if !ui.References[0].Resolved || ui.References[0].Resolution != "heading" {
		t.Fatalf("reference resolution = resolved:%v kind:%q error:%q, want resolved heading", ui.References[0].Resolved, ui.References[0].Resolution, ui.References[0].ResolutionError)
	}
}

func writeMapFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func featureMapByDirName(t *testing.T, featureMaps []FeatureMap, dirName string) FeatureMap {
	t.Helper()
	for _, featureMap := range featureMaps {
		if featureMap.Feature.DirName == dirName {
			return featureMap
		}
	}

	t.Fatalf("feature map %q not found", dirName)
	return FeatureMap{}
}
