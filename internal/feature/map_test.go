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

- builds on: 0002-beta
- depends on: 9999-missing-feature
`)
	writeMapFile(t, filepath.Join(betaDir, "SPEC.md"), `# SPEC

## RELATIONSHIPS

none
`)

	projectMap, err := BuildProjectMap(projectRoot, config.Default())
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}
	if len(projectMap.Features) != 2 {
		t.Fatalf("BuildProjectMap() feature count = %d, want 2", len(projectMap.Features))
	}

	alpha := projectMap.Features[0]
	if alpha.Feature.DirName != "0001-alpha" {
		t.Fatalf("first feature = %s, want 0001-alpha", alpha.Feature.DirName)
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

	beta := projectMap.Features[1]
	if len(beta.Incoming) != 1 {
		t.Fatalf("beta incoming len = %d, want 1", len(beta.Incoming))
	}
	if beta.Incoming[0].SourceFeatureID != "0001-alpha" {
		t.Fatalf("beta incoming source = %s, want 0001-alpha", beta.Incoming[0].SourceFeatureID)
	}
}

func TestBuildProjectMap_IncludesDocumentMetadata(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "0001-alpha"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writeMapFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "BRAINSTORM.md"), "# BRAINSTORM\n\n## RELATIONSHIPS\n\nnone\n")
	writeMapFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")

	projectMap, err := BuildProjectMap(projectRoot, config.Default())
	if err != nil {
		t.Fatalf("BuildProjectMap() error = %v", err)
	}

	if len(projectMap.GlobalDocuments) != 2 {
		t.Fatalf("global document count = %d, want 2", len(projectMap.GlobalDocuments))
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

func writeMapFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
