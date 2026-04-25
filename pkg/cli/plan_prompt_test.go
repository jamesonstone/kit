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
	feat := &feature.Feature{Slug: "sample", DirName: "0012-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, true)
		if err != nil {
			t.Fatalf("outputStandardPlanPrompt() error = %v", err)
		}
	})

	checks := []string{
		"Populate or refresh the `## DEPENDENCIES` table",
		"`Status` must be one of `active`, `optional`, or `stale`",
		"Use an RLM-style just-in-time prior-work pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"- DEPENDENCIES",
		"the ## DEPENDENCIES section must be current before sign-off",
		"no section in `PLAN.md` may remain empty or contain only an HTML TODO comment",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
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
	feat := &feature.Feature{Slug: "sample", DirName: "0012-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, true)
		if err != nil {
			t.Fatalf("outputWarpPlanPrompt() error = %v", err)
		}
	})

	checks := []string{
		"Populate or refresh the `## DEPENDENCIES` table",
		"DEPENDENCIES: the resources that shape the implementation strategy",
		"Use an RLM-style just-in-time prior-work pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"the ## DEPENDENCIES section must be current before sign-off",
		"no section in `PLAN.md` may remain empty or contain only an HTML TODO comment",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}
