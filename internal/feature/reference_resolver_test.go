package feature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestResolveReference_CompositeTargets(t *testing.T) {
	projectRoot := t.TempDir()
	writeResolverFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeResolverFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeResolverFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	resolution := resolveReference(projectRoot, config.Default(), document.MetadataReference{
		Name:       "Instruction files",
		Type:       "doc",
		Target:     "AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md",
		Relation:   document.ReferenceRelationInforms,
		ReadPolicy: document.ReferenceReadPolicyConditional,
		UsedFor:    "instruction routing",
		Status:     document.ReferenceStatusActive,
	})

	if !resolution.Resolved || resolution.Kind != "composite" {
		t.Fatalf("resolution = %#v, want resolved composite", resolution)
	}
}

func TestResolveReference_GlobTarget(t *testing.T) {
	projectRoot := t.TempDir()
	writeResolverFile(t, filepath.Join(projectRoot, "pkg", "cli", "map_test.go"), "package cli\n")

	resolution := resolveReference(projectRoot, config.Default(), document.MetadataReference{
		Name:       "Prompt tests",
		Type:       "tests",
		Target:     "pkg/cli/*test.go",
		Relation:   document.ReferenceRelationVerifies,
		ReadPolicy: document.ReferenceReadPolicyEvidence,
		UsedFor:    "test coverage",
		Status:     document.ReferenceStatusActive,
	})

	if !resolution.Resolved {
		t.Fatalf("resolution = %#v, want resolved glob", resolution)
	}
}

func TestResolveReference_LogicalTargets(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cases := []document.MetadataReference{
		{
			Name:       "Cobra",
			Type:       "library",
			Target:     "github.com/spf13/cobra",
			Relation:   document.ReferenceRelationUses,
			ReadPolicy: document.ReferenceReadPolicyConditional,
			UsedFor:    "CLI behavior",
			Status:     document.ReferenceStatusActive,
		},
		{
			Name:       "Git CLI",
			Type:       "tool",
			Target:     "git rev-parse --git-common-dir",
			Relation:   document.ReferenceRelationUses,
			ReadPolicy: document.ReferenceReadPolicyConditional,
			UsedFor:    "worktree coordination",
			Status:     document.ReferenceStatusActive,
		},
		{
			Name:       "Frontend profile",
			Type:       "profile",
			Target:     "--profile=frontend",
			Relation:   document.ReferenceRelationGuides,
			ReadPolicy: document.ReferenceReadPolicyConditional,
			UsedFor:    "frontend prompts",
			Status:     document.ReferenceStatusActive,
		},
	}

	for _, tc := range cases {
		resolution := resolveReference(projectRoot, cfg, tc)
		if !resolution.Resolved {
			t.Fatalf("%s resolution = %#v, want resolved", tc.Name, resolution)
		}
	}
}

func writeResolverFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
