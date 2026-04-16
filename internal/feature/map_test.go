package feature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestBuildProjectMap_CollectsRelationshipsAndIncomingEdges(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")

	alphaDir := filepath.Join(specsDir, "0001-alpha")
	betaDir := filepath.Join(specsDir, "0002-beta")
	if err := os.MkdirAll(alphaDir, 0755); err != nil {
		t.Fatalf("MkdirAll(alpha) error = %v", err)
	}
	if err := os.MkdirAll(betaDir, 0755); err != nil {
		t.Fatalf("MkdirAll(beta) error = %v", err)
	}

	writeMapFile(t, filepath.Join(alphaDir, "SPEC.md"), `# SPEC

## RELATIONSHIPS

- builds on: `+"`0002-beta`"+`
- depends on: 9999-missing-feature
`)
	writeMapFile(t, filepath.Join(betaDir, "SPEC.md"), `# SPEC

## RELATIONSHIPS

none
`)

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}
	if len(projectMap.Features) != 2 {
		t.Fatalf("BuildProjectMap() feature count = %d, want 2", len(projectMap.Features))
	}

	alpha := featureMapByDirName(t, projectMap.Features, "0001-alpha")
	if alpha.Feature.DirName != "0001-alpha" {
		t.Fatalf("alpha feature = %s, want 0001-alpha", alpha.Feature.DirName)
	}
	if len(alpha.Outgoing) != 2 {
		t.Fatalf("alpha outgoing len = %d, want 2", len(alpha.Outgoing))
	}
	if !alpha.Outgoing[0].Resolved {
		t.Fatalf("expected resolved relationship for %s", alpha.Outgoing[0].TargetFeatureID)
	}
	if alpha.Outgoing[1].Resolved {
		t.Fatalf("expected unresolved relationship for %s", alpha.Outgoing[1].TargetFeatureID)
	}

	beta := featureMapByDirName(t, projectMap.Features, "0002-beta")
	if len(beta.Incoming) != 1 {
		t.Fatalf("beta incoming len = %d, want 1", len(beta.Incoming))
	}
	if beta.Incoming[0].SourceFeatureID != "0001-alpha" {
		t.Fatalf("beta incoming source = %s, want 0001-alpha", beta.Incoming[0].SourceFeatureID)
	}
}

func TestBuildProjectMap_KeepsValidEdgesAndWarnsOnMalformedRelationshipLines(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "0001-alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), `# SPEC

## RELATIONSHIPS

- builds on: 0002-beta
- follows: 0003-gamma
`)
	writeMapFile(t, filepath.Join(specsDir, "0002-beta", "SPEC.md"), `# SPEC

## RELATIONSHIPS

none
`)

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}

	if len(projectMap.Features) != 2 {
		t.Fatalf("BuildProjectMap() feature count = %d, want 2", len(projectMap.Features))
	}
	alpha := featureMapByDirName(t, projectMap.Features, "0001-alpha")
	if len(alpha.Outgoing) != 1 {
		t.Fatalf("alpha outgoing len = %d, want 1 valid edge", len(alpha.Outgoing))
	}
	if len(projectMap.Warnings) != 1 {
		t.Fatalf("map warnings len = %d, want 1", len(projectMap.Warnings))
	}
	if projectMap.Warnings[0].FeatureID != "0001-alpha" {
		t.Fatalf("warning feature = %s, want 0001-alpha", projectMap.Warnings[0].FeatureID)
	}
	if projectMap.Warnings[0].Document != "SPEC.md" {
		t.Fatalf("warning document = %s, want SPEC.md", projectMap.Warnings[0].Document)
	}
	if projectMap.Warnings[0].Line != "- follows: 0003-gamma" {
		t.Fatalf("warning line = %q, want invalid line", projectMap.Warnings[0].Line)
	}
}

func TestBuildProjectMap_IncludesDocumentMetadata(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "0001-alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "BRAINSTORM.md"), "# BRAINSTORM\n\n## RELATIONSHIPS\n\nnone\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}

	if len(projectMap.GlobalDocuments) != 5 {
		t.Fatalf("global document count = %d, want 5", len(projectMap.GlobalDocuments))
	}
	if len(projectMap.Features) != 1 {
		t.Fatalf("feature count = %d, want 1", len(projectMap.Features))
	}

	docs := projectMap.Features[0].Documents
	if len(docs) != 5 {
		t.Fatalf("document count = %d, want 5", len(docs))
	}
	if !docs[0].Exists || docs[0].Required {
		t.Fatalf("brainstorm metadata = %+v, want optional present document", docs[0])
	}
	if !docs[1].Exists || !docs[1].Required {
		t.Fatalf("spec metadata = %+v, want required present document", docs[1])
	}
	if docs[2].Exists || !docs[2].Required {
		t.Fatalf("plan metadata = %+v, want required missing document", docs[2])
	}
}

func TestBuildProjectMap_IncludesTOCGlobalDocuments(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "0001-alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeMapFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeMapFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "agents", "README.md"), "# Agents Docs\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "references", "README.md"), "# References\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}

	globalPaths := make([]string, 0, len(projectMap.GlobalDocuments))
	for _, doc := range projectMap.GlobalDocuments {
		globalPaths = append(globalPaths, doc.Path)
	}

	for _, expected := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		".github/copilot-instructions.md",
		"docs/agents/README.md",
		"docs/references/README.md",
	} {
		found := false
		for _, path := range globalPaths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected global documents to include %q, got %v", expected, globalPaths)
		}
	}
}

func TestBuildProjectMap_UsesDependencyOrderForProjectGraph(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	for _, dirName := range []string{"0001-ui", "0002-auth", "0003-api"} {
		if err := os.MkdirAll(filepath.Join(specsDir, dirName), 0755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dirName, err)
		}
	}

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-ui", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\n- depends on: 0003-api\n")
	writeMapFile(t, filepath.Join(specsDir, "0002-auth", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")
	writeMapFile(t, filepath.Join(specsDir, "0003-api", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\n- builds on: 0002-auth\n")

	projectMap, err := BuildProjectMap(projectRoot, cfg)
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}

	got := []string{
		projectMap.Features[0].Feature.DirName,
		projectMap.Features[1].Feature.DirName,
		projectMap.Features[2].Feature.DirName,
	}
	want := []string{"0002-auth", "0003-api", "0001-ui"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("feature order[%d] = %s, want %s (full order %v)", i, got[i], want[i], got)
		}
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
