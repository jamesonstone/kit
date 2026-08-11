package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeDocumentsV2BoundaryAndPrimaryFlow(t *testing.T) {
	content := readRepositoryFile(t, "README.md")
	for _, required := range []string{
		"## Major Update",
		"Kit 2.0 intentionally removes",
		"kit reconcile --include-files --dry-run --diff",
		"kit capabilities context resolve --json",
		"kit spec my-feature",
		"kit context resolve --workflow implementation-delivery --feature my-feature --json",
		"kit usage disable --global",
		"never records command arguments",
		"365 days",
		"16 MiB total",
		"The separate `git-wt` binary remains available and unchanged",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("README missing %q", required)
		}
	}
	if strings.Index(content, "kit spec my-feature") > strings.Index(content, "kit context resolve --workflow implementation-delivery") {
		t.Error("README resolves a feature before creating or adopting its spec")
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

func TestReleaseWorkflowEstablishesV2ThenResumesPatchBumps(t *testing.T) {
	content := readRepositoryFile(t, ".github/workflows/release-tag-main.yml")
	for _, required := range []string{
		`next_tag="v2.0.0"`,
		"if (( major < 2 )); then",
		`next_tag="v${major}.${minor}.$((patch + 1))"`,
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
