package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeDocumentsV3BoundaryAndPrimaryFlow(t *testing.T) {
	content := readRepositoryFile(t, "README.md")
	for _, required := range []string{
		"## Major Update",
		"Kit 3.0 makes subagent orchestration capability-aware",
		"github.com/jamesonstone/kit/v3/cmd/kit@latest",
		"docs/migration-v3.md",
		"kit reconcile --include-files --dry-run --diff",
		"kit capabilities context resolve --json",
		"kit spec my-feature",
		"kit context resolve --workflow implementation-delivery --feature my-feature --json",
		"`release-orchestration`",
		"`kit pr orchestrate`",
		"kit usage disable --global",
		"never records command arguments",
		"365 days",
		"16 MiB total",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("README missing %q", required)
		}
	}
	if strings.Index(content, "kit spec my-feature") > strings.Index(content, "kit context resolve --workflow implementation-delivery") {
		t.Error("README resolves a feature before creating or adopting its spec")
	}
	for _, removed := range []string{"git-wt", "git wt"} {
		if strings.Contains(content, removed) {
			t.Errorf("README still documents removed command %q", removed)
		}
	}
}

func TestMigrationV3DocumentsBreakingBoundary(t *testing.T) {
	content := readRepositoryFile(t, "docs/migration-v3.md")
	for _, required := range []string{
		"github.com/jamesonstone/kit/v3",
		"`--max-subagents`",
		"`--single-agent`",
		"exactly three default instruction targets",
		"Logical roles, plans, task lists",
		"Historical v1 and v2 specifications",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("v3 migration guide missing %q", required)
		}
	}
}

func TestMigrationDocumentsWeeklyHealthCompatibilityBoundary(t *testing.T) {
	content := readRepositoryFile(t, "docs/migration-v2.md")
	for _, required := range []string{
		"historical specifications",
		"`kit reconcile` has not been redesigned",
		"kit capabilities usage --json",
		"kit usage status --json",
		"kit usage report --since 90d --json",
		"Do not update the automation before the released v2 binary is installed",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("migration guide missing %q", required)
		}
	}
}

func TestReleaseWorkflowEstablishesV3ThenResumesPatchBumps(t *testing.T) {
	content := readRepositoryFile(t, ".github/workflows/release-tag-main.yml")
	for _, required := range []string{
		"queue: max",
		".github/scripts/release-next-tag.sh HEAD",
		"uses: jamesonstone/mint@v0.2.1",
		"command: release-tag",
		"command: github-release",
		"release-push: \"true\"",
		"needs.prepare-release.outputs.next_tag != ''",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow missing %q", required)
		}
	}
}

func readRepositoryFile(t *testing.T, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
