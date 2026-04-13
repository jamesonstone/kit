package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestBuildCatchupPrompt(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0007-catchup-command")
	feat := &feature.Feature{
		Slug:    "catchup-command",
		DirName: "0007-catchup-command",
		Path:    featurePath,
	}
	status := &feature.FeatureStatus{
		ID:      "0007",
		Name:    feat.Slug,
		Path:    featurePath,
		Summary: "Resume a feature safely before implementation restarts.",
		Phase:   feature.PhaseReflect,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: filepath.Join(featurePath, "BRAINSTORM.md")},
			"spec":       {Exists: true, Path: filepath.Join(featurePath, "SPEC.md")},
			"plan":       {Exists: true, Path: filepath.Join(featurePath, "PLAN.md")},
			"tasks":      {Exists: true, Path: filepath.Join(featurePath, "TASKS.md")},
		},
		Progress: &feature.TaskProgress{Total: 10, Complete: 7},
	}

	prompt := buildCatchupPrompt(feat, status, projectRoot)

	checks := []string{
		"/plan",
		"Catch up on feature: catchup-command",
		"Current stage: reflect",
		"Current state:",
		"task progress 7/10 complete",
		"CONSTITUTION.md",
		"PROJECT_PROGRESS_SUMMARY.md",
		"BRAINSTORM.md",
		"SPEC.md",
		"PLAN.md",
		"TASKS.md",
		"Stay in plan mode",
		"Start by asking clarifying questions",
		"Do NOT switch from catch-up/planning into implementation until the user explicitly approves that move",
		"`kit summarize catchup-command`",
		"do not duplicate the full `kit handoff` workflow",
		"do not output implementation instructions like `kit implement` unless the user explicitly asks to proceed",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "/plan\n\nCatch up on feature: catchup-command\n\n") {
		t.Fatalf("expected prompt to start with /plan catchup header, got %q", prompt[:48])
	}
}

func TestBuildCatchupPromptForCompleteFeature(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0007-catchup-command")
	feat := &feature.Feature{
		Slug:    "catchup-command",
		DirName: "0007-catchup-command",
		Path:    featurePath,
	}
	status := &feature.FeatureStatus{
		ID:    "0007",
		Name:  feat.Slug,
		Path:  featurePath,
		Phase: feature.PhaseComplete,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: filepath.Join(featurePath, "BRAINSTORM.md")},
			"spec":       {Exists: true, Path: filepath.Join(featurePath, "SPEC.md")},
			"plan":       {Exists: true, Path: filepath.Join(featurePath, "PLAN.md")},
			"tasks":      {Exists: true, Path: filepath.Join(featurePath, "TASKS.md")},
		},
		Progress: &feature.TaskProgress{Total: 3, Complete: 3},
	}

	prompt := buildCatchupPrompt(feat, status, projectRoot)
	checks := []string{
		"Current stage: complete",
		"This feature is already marked `complete`; treat this catch-up as review or reopen triage only",
		"Do not assume implementation should resume unless the user explicitly asks to reopen work on this feature",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected complete-phase prompt to contain %q", check)
		}
	}
}

func TestBuildCatchupPrompt_IncludesRepoDocsForTOCRepos(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, "docs", "agents", "README.md"), "# Agents Docs\n")
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "README.md"), "# References\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0007-catchup-command")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	feat := &feature.Feature{
		Slug:    "catchup-command",
		DirName: "0007-catchup-command",
		Path:    featurePath,
	}
	status := &feature.FeatureStatus{
		ID:    "0007",
		Name:  feat.Slug,
		Path:  featurePath,
		Phase: feature.PhasePlan,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: filepath.Join(featurePath, "BRAINSTORM.md")},
			"spec":       {Exists: true, Path: filepath.Join(featurePath, "SPEC.md")},
			"plan":       {Exists: true, Path: filepath.Join(featurePath, "PLAN.md")},
			"tasks":      {Exists: false, Path: filepath.Join(featurePath, "TASKS.md")},
		},
	}

	prompt := buildCatchupPrompt(feat, status, projectRoot)
	for _, check := range []string{
		"AGENTS DOCS",
		"REFERENCES",
		"`docs/agents/README.md`",
		"`docs/references/README.md`",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
}
