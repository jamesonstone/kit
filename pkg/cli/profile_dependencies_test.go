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
		name   string
		target string
		status string
		want   bool
	}{
		{
			name:   "active canonical reference",
			target: frontendProfileReferenceTarget,
			status: document.ReferenceStatusActive,
			want:   true,
		},
		{
			name:   "optional reference does not activate",
			target: frontendProfileReferenceTarget,
			status: document.ReferenceStatusOptional,
		},
		{
			name:   "stale reference does not activate",
			target: frontendProfileReferenceTarget,
			status: document.ReferenceStatusStale,
		},
		{
			name:   "wrong target does not activate",
			target: "docs/agents/frontend.md",
			status: document.ReferenceStatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			featurePath := filepath.Join(t.TempDir(), "docs", "specs", "0001-ui")
			writeFile(t, filepath.Join(featurePath, "SPEC.md"), frontendProfileReferenceDoc(tt.target, tt.status))

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
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), frontendProfileReferenceDoc(frontendProfileReferenceTarget, document.ReferenceStatusActive))

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
		"target: docs/notes/0001-ui/design",
		"<!-- keep this comment -->",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected updated dependency table to contain %q, got:\n%s", check, text)
		}
	}
	doc := document.Parse(text, specPath, document.TypeSpec)
	if !hasReference(doc.References(), frontendProfileDependencyName, frontendProfileReferenceTarget, document.ReferenceStatusActive) {
		t.Fatalf("expected one active frontend profile reference in front matter, got %#v", doc.References())
	}
	if hasReference(doc.References(), "Existing API", "https://example.test", document.ReferenceStatusActive) {
		t.Fatalf("expected legacy body dependency not to be carried into front matter, got %#v", doc.References())
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
	if !hasReferenceWithUsedFor(doc.References(), frontendProfileDependencyName, frontendProfileReferenceTarget, "apply frontend-specific coding-agent instruction set") {
		t.Fatalf("expected frontend profile reference wording to be refreshed in front matter, got %#v", doc.References())
	}
	if !hasReferenceWithUsedFor(doc.References(), designMaterialsDependencyName, "docs/notes/0001-ui/design", "optional frontend design input") {
		t.Fatalf("expected design reference wording to be refreshed in front matter, got %#v", doc.References())
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

func hasReference(references []document.MetadataReference, name, target, status string) bool {
	for _, reference := range references {
		if reference.Name == name && reference.Target == target && reference.Status == status {
			return true
		}
	}
	return false
}

func hasReferenceWithUsedFor(references []document.MetadataReference, name, target, usedFor string) bool {
	for _, reference := range references {
		if reference.Name == name && reference.Target == target && reference.UsedFor == usedFor {
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

func frontendProfileReferenceDoc(target, status string) string {
	return `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: ui
  dir: 0001-ui
references:
  - name: Frontend profile
    type: profile
    target: ` + target + `
    relation: guides
    read_policy: conditional
    used_for: apply frontend-specific coding-agent instruction set
    status: ` + status + `
---
# SPEC

## DEPENDENCIES

References are tracked in front matter.
`
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
