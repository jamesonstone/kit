package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeDocumentsCurrentSurface(t *testing.T) {
	content := readRepositoryFile(t, "README.md")
	for _, required := range []string{
		"github.com/jamesonstone/kit/v3/cmd/kit@latest",
		"docs/migration-v3.md",
		"kit reconcile --dry-run --diff",
		"kit spec my-feature",
		"kit usage disable --global",
		"never records command arguments",
		"365 days",
		"16 MiB total",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("README missing %q", required)
		}
	}
	for _, removed := range []string{"git-wt", "git wt", "kit context resolve", "kit capabilities", "kit dispatch", "kit pr fix", "kit pr orchestrate", "kit improve", "kit instructions"} {
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
		"uses: jamesonstone/mint@b97969136d5a43d0982c46c6f185868db16d14bf # v0.5.0",
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
