package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func TestFeatureHasActiveFrontendProfileDependency(t *testing.T) {
	tests := []struct {
		name string
		row  string
		want bool
	}{
		{
			name: "active canonical row",
			row:  "| `Frontend profile` | `profile` | `--profile=frontend` | apply frontend-specific coding-agent instruction set | Active |",
			want: true,
		},
		{
			name: "optional row does not activate",
			row:  "| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | optional |",
		},
		{
			name: "stale row does not activate",
			row:  "| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | stale |",
		},
		{
			name: "wrong location does not activate",
			row:  "| Frontend profile | profile | docs/agents/frontend.md | apply frontend-specific coding-agent instruction set | active |",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
			writeFile(t, filepath.Join(featurePath, "SPEC.md"), dependencyDoc(tt.row))

			got := featureHasActiveFrontendProfileDependency(featurePath)
			if got != tt.want {
				t.Fatalf("featureHasActiveFrontendProfileDependency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeatureHasActiveFrontendProfileDependencyIgnoresMalformedTables(t *testing.T) {
	featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), `# SPEC

## DEPENDENCIES

| Dependency | Type | Location |
| ---------- | ---- | -------- |
| Frontend profile | profile | --profile=frontend |
`)

	if featureHasActiveFrontendProfileDependency(featurePath) {
		t.Fatal("expected malformed dependency table not to activate frontend profile")
	}
}

func TestEffectivePromptProfileResolution(t *testing.T) {
	featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), dependencyDoc("| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |"))

	restorePromptProfileState(t, promptProfileNone, false)
	if got := effectivePromptProfile(featurePath); got != promptProfileFrontend {
		t.Fatalf("effectivePromptProfile(active feature) = %q, want %q", got, promptProfileFrontend)
	}

	restorePromptProfileState(t, promptProfileNone, true)
	if got := effectivePromptProfile(featurePath); got != promptProfileNone {
		t.Fatalf("effectivePromptProfile(explicit empty) = %q, want empty profile", got)
	}

	restorePromptProfileState(t, promptProfileFrontend, true)
	if got := effectivePromptProfile(""); got != promptProfileFrontend {
		t.Fatalf("effectivePromptProfile(explicit frontend) = %q, want %q", got, promptProfileFrontend)
	}
}

func TestEnsureFrontendProfileDependencyRowsAppendsIdempotently(t *testing.T) {
	featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, `# SPEC

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |
| Frontend profile | profile | --profile=frontend | previous profile experiment | stale |
| Existing API | api | https://example.test | prior input | active |

<!-- keep this comment -->
`)

	changed, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, "0001-ui")
	if err != nil {
		t.Fatalf("ensureFrontendProfileDependencyRows() error = %v", err)
	}
	if !changed {
		t.Fatal("expected dependency rows to be appended")
	}

	changed, err = ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, "0001-ui")
	if err != nil {
		t.Fatalf("second ensureFrontendProfileDependencyRows() error = %v", err)
	}
	if changed {
		t.Fatal("expected second dependency ensure to be a no-op")
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"| Frontend profile | profile | --profile=frontend | previous profile experiment | stale |",
		"| Existing API | api | https://example.test | prior input | active |",
		"| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |",
		"| Design materials | design | docs/notes/0001-ui/design | optional frontend design input | optional |",
		"<!-- keep this comment -->",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected updated dependency table to contain %q, got:\n%s", check, text)
		}
	}
	if strings.Contains(text, "| none | n/a | n/a |") {
		t.Fatalf("expected placeholder none row to be removed, got:\n%s", text)
	}
	if count := strings.Count(text, "| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |"); count != 1 {
		t.Fatalf("expected one active frontend profile row, got %d in:\n%s", count, text)
	}
}

func TestEnsureFrontendProfileDependencyRowsRefreshesCanonicalRows(t *testing.T) {
	featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, `# SPEC

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Frontend profile | profile | --profile=frontend | old wording | active |
| Design materials | design | docs/notes/0001-ui/design | old design wording | optional |
`)

	changed, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, "0001-ui")
	if err != nil {
		t.Fatalf("ensureFrontendProfileDependencyRows() error = %v", err)
	}
	if !changed {
		t.Fatal("expected dependency rows to be refreshed")
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	if strings.Contains(text, "old wording") || strings.Contains(text, "old design wording") {
		t.Fatalf("expected old canonical row wording to be refreshed, got:\n%s", text)
	}
	for _, check := range []string{
		"| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |",
		"| Design materials | design | docs/notes/0001-ui/design | optional frontend design input | optional |",
	} {
		if count := strings.Count(text, check); count != 1 {
			t.Fatalf("expected one refreshed row %q, got %d in:\n%s", check, count, text)
		}
	}
}

func dependencyDoc(row string) string {
	return `# SPEC

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
` + row + "\n"
}

func restorePromptProfileState(t *testing.T, profile promptProfile, explicit bool) {
	t.Helper()
	previous := selectedPromptProfile
	previousExplicit := selectedPromptProfileExplicit
	selectedPromptProfile = profile
	selectedPromptProfileExplicit = explicit
	t.Cleanup(func() {
		selectedPromptProfile = previous
		selectedPromptProfileExplicit = previousExplicit
	})
}
