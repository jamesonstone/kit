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
		"name: Frontend profile",
		"used_for: apply frontend-specific coding-agent instruction set",
		"name: Design materials",
		"location: docs/notes/0001-ui/design",
		"<!-- keep this comment -->",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected updated dependency table to contain %q, got:\n%s", check, text)
		}
	}
	doc := document.Parse(text, specPath, document.TypeSpec)
	if !hasDependency(doc.Dependencies(), frontendProfileDependencyName, frontendProfileDependencyLocation, document.DependencyStatusActive) {
		t.Fatalf("expected one active frontend profile dependency in front matter, got %#v", doc.Dependencies())
	}
	if !hasDependency(doc.Dependencies(), "Existing API", "https://example.test", document.DependencyStatusActive) {
		t.Fatalf("expected existing legacy dependency to be carried into front matter, got %#v", doc.Dependencies())
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
	doc := document.Parse(text, specPath, document.TypeSpec)
	if !hasDependencyWithUsedFor(doc.Dependencies(), frontendProfileDependencyName, frontendProfileDependencyLocation, "apply frontend-specific coding-agent instruction set") {
		t.Fatalf("expected frontend profile dependency wording to be refreshed in front matter, got %#v", doc.Dependencies())
	}
	if !hasDependencyWithUsedFor(doc.Dependencies(), designMaterialsDependencyName, "docs/notes/0001-ui/design", "optional frontend design input") {
		t.Fatalf("expected design dependency wording to be refreshed in front matter, got %#v", doc.Dependencies())
	}
}

func TestEnsureFrontendProfileDependencyRowsErrorsOnMalformedFrontMatter(t *testing.T) {
	featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: ui
  dir: 0001-ui
# SPEC
`)

	changed, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, "0001-ui")
	if err == nil {
		t.Fatal("ensureFrontendProfileDependencyRows() error = nil, want malformed front matter error")
	}
	if changed {
		t.Fatal("ensureFrontendProfileDependencyRows() changed = true, want false")
	}
}

func hasDependency(dependencies []document.MetadataDependency, name, location, status string) bool {
	for _, dependency := range dependencies {
		if dependency.Name == name && dependency.Location == location && dependency.Status == status {
			return true
		}
	}
	return false
}

func hasDependencyWithUsedFor(dependencies []document.MetadataDependency, name, location, usedFor string) bool {
	for _, dependency := range dependencies {
		if dependency.Name == name && dependency.Location == location && dependency.UsedFor == usedFor {
			return true
		}
	}
	return false
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
