package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestFeaturePromptCommandsExposePromptOnlyFlag(t *testing.T) {
	commandNames := []string{
		"brainstorm",
		"spec",
		"plan",
		"tasks",
		"implement",
		"reflect",
		"reconcile",
		"catchup",
		"handoff",
	}

	for _, name := range commandNames {
		cmd, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("rootCmd.Find(%q) error = %v", name, err)
		}
		if cmd.Flags().Lookup("prompt-only") == nil {
			t.Fatalf("expected %q to expose --prompt-only", name)
		}
	}

	skillMineCmd, _, err := rootCmd.Find([]string{"skill", "mine"})
	if err != nil {
		t.Fatalf("rootCmd.Find(skill mine) error = %v", err)
	}
	if skillMineCmd.Flags().Lookup("prompt-only") == nil {
		t.Fatalf("expected skill mine to expose --prompt-only")
	}
}

func TestRunSpecPromptOnly_RequiresExistingSpec(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "BRAINSTORM.md"), "# BRAINSTORM\n")

	err := runSpecPromptOnly([]string{"sample"}, projectRoot, config.Default(), true)
	if err == nil || !strings.Contains(err.Error(), "SPEC.md not found") {
		t.Fatalf("expected missing SPEC.md error, got %v", err)
	}
}

func TestRunPlanPromptOnly_RequiresExistingPlan(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "SPEC.md"), "# SPEC\n")

	err := runPlanPromptOnly([]string{"sample"}, projectRoot, config.Default(), false, true)
	if err == nil || !strings.Contains(err.Error(), "PLAN.md not found") {
		t.Fatalf("expected missing PLAN.md error, got %v", err)
	}
}

func TestRunTasksPromptOnly_RequiresExistingTasks(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "SPEC.md"), "# SPEC\n")
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "PLAN.md"), "# PLAN\n")

	err := runTasksPromptOnly([]string{"sample"}, projectRoot, config.Default(), true)
	if err == nil || !strings.Contains(err.Error(), "TASKS.md not found") {
		t.Fatalf("expected missing TASKS.md error, got %v", err)
	}
}
