package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestInitRefreshMigratesExactGeneratedV2InstructionsToV3(t *testing.T) {
	projectRoot := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistry(t)
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatal(err)
	}
	for _, path := range instructionArtifactPaths(cfg, instructionFileSelection{}, config.InstructionScaffoldVersionTOC, true) {
		content, _, err := instructionArtifactContent(path, config.InstructionScaffoldVersionTOC)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(projectRoot, path), content)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		if err := runInit(initCmd, nil); err != nil {
			t.Fatalf("runInit() error = %v", err)
		}
	})
	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if updated.InstructionScaffoldVersion != config.InstructionScaffoldVersionMemory {
		t.Fatalf("instruction version = %d, want 3", updated.InstructionScaffoldVersion)
	}
	if got := readFile(t, filepath.Join(projectRoot, agentsMDPath)); got != templates.MemoryAgentsMD {
		t.Fatalf("AGENTS.md was not migrated to V3:\n%s", got)
	}
}

func TestInitRefreshPreservesCustomizedV2Instructions(t *testing.T) {
	projectRoot := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistry(t)
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatal(err)
	}
	for _, path := range instructionArtifactPaths(cfg, instructionFileSelection{}, config.InstructionScaffoldVersionTOC, true) {
		content, _, err := instructionArtifactContent(path, config.InstructionScaffoldVersionTOC)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(projectRoot, path), content)
	}
	agentsPath := filepath.Join(projectRoot, agentsMDPath)
	custom := readFile(t, agentsPath) + "\n## Local Policy\n\nkeep me\n"
	writeFile(t, agentsPath, custom)

	plan, err := buildInitRefreshPlan(t.Context(), projectRoot, initRefreshOptions{outputOnly: true, force: true})
	if err != nil {
		t.Fatal(err)
	}
	if plan.cfg.InstructionScaffoldVersion != config.InstructionScaffoldVersionTOC {
		t.Fatalf("planned instruction version = %d, want preserved V2", plan.cfg.InstructionScaffoldVersion)
	}
	if !strings.Contains(strings.Join(plan.notes, " "), "customized V2 instruction artifacts were preserved") {
		t.Fatalf("migration advisory missing: %#v", plan.notes)
	}
	if got := readFile(t, agentsPath); got != custom {
		t.Fatalf("planning mutated customized instructions:\n%s", got)
	}
}

func TestApplyInitRefreshFileChangesAtomicallyRollsBack(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.md")
	writeFile(t, first, "before\n")
	blocker := filepath.Join(root, "blocker")
	writeFile(t, blocker, "not a directory\n")
	changes := []initRefreshFileChange{
		*newInitRefreshFileChange(root, "first.md", "before\n", "after\n", instructionFileUpdated),
		*newInitRefreshFileChange(root, filepath.Join("blocker", "child.md"), "", "child\n", instructionFileCreated),
	}
	if err := applyInitRefreshFileChangesAtomically(changes); err == nil {
		t.Fatal("expected atomic apply failure")
	}
	data, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "before\n" {
		t.Fatalf("first file was not rolled back: %q", data)
	}
}
