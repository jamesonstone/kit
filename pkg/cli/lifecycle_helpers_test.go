package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func setupLifecycleTestProject(t *testing.T) (string, *config.Config) {
	t.Helper()

	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), readyConstitutionForTest())
	for _, relativePath := range instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		cfg.InstructionScaffoldVersion,
		true,
	) {
		content, _, err := instructionArtifactContent(relativePath, cfg.InstructionScaffoldVersion)
		if err != nil {
			t.Fatalf("instructionArtifactContent(%q) error = %v", relativePath, err)
		}
		writeFile(t, filepath.Join(projectRoot, relativePath), content)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "docs", "specs"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	return projectRoot, cfg
}

func readyConstitutionForTest() string {
	return `# CONSTITUTION

## PRINCIPLES

Clarity and correctness guide all project decisions.

## CONSTRAINTS

Keep generated and human-authored project guidance aligned.

## NON-GOALS

No additional non-goals for this test fixture.

## DEFINITIONS

Kit-managed project means a repository with Kit configuration and guidance.
`
}

func setRemoveFlagsForTest(t *testing.T, yes bool, notes bool) {
	t.Helper()
	previousYes := removeYes
	previousNotes := removeNotes
	removeYes = yes
	removeNotes = notes
	t.Cleanup(func() {
		removeYes = previousYes
		removeNotes = previousNotes
	})
}

func useTestStdin(t *testing.T, input string) {
	t.Helper()

	originalStdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		os.Stdin = originalStdin
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close() error = %v", err)
		}
	})

	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	os.Stdin = reader
}

var featureRefAlpha = feature.Feature{
	DirName: "0001-alpha",
	Path:    "/tmp/0001-alpha",
}

func pausedStatus() *feature.FeatureStatus {
	return &feature.FeatureStatus{
		ID:     "0001",
		Name:   "alpha",
		Phase:  feature.PhaseSpec,
		Paused: true,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: false, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}
}
