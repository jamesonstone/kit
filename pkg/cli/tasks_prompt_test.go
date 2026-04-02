package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
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
		"no section in `TASKS.md` may remain empty or contain only an HTML TODO comment",
		"if there are no blockers or ordering notes, replace placeholder comments with \"no additional information required\" or \"not applicable\"",
		"otherwise write \"not required\"",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}
