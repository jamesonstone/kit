package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestOutputStandardPlanPrompt_IncludesDependencyGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	planPath := filepath.Join(featurePath, "PLAN.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")

	writeFile(t, specPath, "# SPEC\n")
	writeFile(t, planPath, "# PLAN\n")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	cfg.ConstitutionPath = filepath.Join("governance", "PROJECT_RULES.md")
	feat := &feature.Feature{Slug: "sample", DirName: "0012-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, true)
		if err != nil {
			t.Fatalf("outputStandardPlanPrompt() error = %v", err)
		}
	})

	checks := []string{
		filepath.Join(projectRoot, cfg.ConstitutionPath),
		"Complete the legacy implementation plan",
		"documentation-only; do not implement product code",
		"Inspect the smallest relevant code, tests, docs, and prior-feature context",
		"Ask concise numbered questions only for a material non-discoverable design choice",
		"simplest viable approach",
		"exact dependencies/references",
		"Map every binding acceptance criterion",
		"validation strategy",
		"`not applicable`",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	if defaultPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md"); strings.Contains(output, defaultPath) {
		t.Fatalf("expected output to use configured constitution path, not %q", defaultPath)
	}
	assertFinalResponseContractHeadings(t, output,
		"Summary",
		"Artifacts Updated",
		"Design Decisions",
		"Implementation Risks",
		"Next Step",
	)
}

func TestOutputWarpPlanPrompt_IncludesDependencyGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	planPath := filepath.Join(featurePath, "PLAN.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")

	writeFile(t, specPath, "# SPEC\n")
	writeFile(t, planPath, "# PLAN\n")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	cfg.ConstitutionPath = filepath.Join("governance", "PROJECT_RULES.md")
	feat := &feature.Feature{Slug: "sample", DirName: "0012-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, true)
		if err != nil {
			t.Fatalf("outputWarpPlanPrompt() error = %v", err)
		}
	})

	checks := []string{
		filepath.Join(projectRoot, cfg.ConstitutionPath),
		"Convert the Warp plan in the current conversation",
		"documentation-only; do not implement product code",
		"SPEC.md wins on conflict",
		"smallest relevant code and test surfaces",
		"Ask concise numbered questions only for a material non-discoverable design choice",
		"simplest viable approach",
		"Map every binding acceptance criterion",
		"introduce no scope beyond SPEC.md",
		"validation strategy",
		"`not applicable`",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	if defaultPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md"); strings.Contains(output, defaultPath) {
		t.Fatalf("expected output to use configured constitution path, not %q", defaultPath)
	}
	for _, obsolete := range []string{
		"additional batches of up to 10 questions",
		"inspect at most 5 prior feature directories",
		"The output of PLAN.md must make TASKS.md obvious and deterministic",
	} {
		if strings.Contains(output, obsolete) {
			t.Fatalf("expected compact Warp prompt to omit %q", obsolete)
		}
	}
	assertFinalResponseContractHeadings(t, output,
		"Summary",
		"Artifacts Updated",
		"Design Decisions",
		"Implementation Risks",
		"Next Step",
	)
}
