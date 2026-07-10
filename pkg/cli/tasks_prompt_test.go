package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestOutputTasksPrompt_IncludesNonEmptySectionGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), "# SPEC\n")
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), "# PLAN\n")
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), "# TASKS\n")

	cfg := config.Default()
	feat := &feature.Feature{Slug: "sample", DirName: "0012-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputTasksPrompt(feat, projectRoot, cfg, true)
		if err != nil {
			t.Fatalf("outputTasksPrompt() error = %v", err)
		}
	})

	checks := []string{
		"Update TASKS.md directly at",
		"do not leave the task breakdown only in chat",
		"Update TASKS.md only; do not implement product code",
		"material non-discoverable decision",
		"stable T001-style IDs",
		"GOAL, SCOPE, ACCEPTANCE, VERIFY, EXPECTED FILES, RISK, ROLLBACK, DEPENDENCIES",
		"binary done condition and required evidence",
		"Update TASKS.md only.",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	if strings.Contains(output, "supporting documentation") {
		t.Fatalf("expected TASKS.md to be the only prompt-authorized mutation, got:\n%s", output)
	}
	assertFinalResponseContractHeadings(t, output,
		"Summary",
		"Artifacts Updated",
		"Task Breakdown",
		"Blocked Items",
		"Next Step",
	)
}

func TestOutputTasksPrompt_IncludesInferredFrontendProfile(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-dashboard")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), frontendProfileReferenceDoc(frontendProfileReferenceTarget, document.ReferenceStatusActive))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), "# PLAN\n")
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), "# TASKS\n")

	restorePromptProfileState(t, promptProfileNone, false)
	cfg := config.Default()
	feat := &feature.Feature{Slug: "dashboard", DirName: "0012-dashboard", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputTasksPrompt(feat, projectRoot, cfg, true)
		if err != nil {
			t.Fatalf("outputTasksPrompt() error = %v", err)
		}
	})

	if !strings.Contains(output, "## Frontend Profile") {
		t.Fatalf("expected inferred frontend profile guidance, got:\n%s", output)
	}
	if !strings.Contains(output, "browser or screenshot evidence") {
		t.Fatalf("expected frontend validation guidance, got:\n%s", output)
	}
}
