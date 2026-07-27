package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func validV3CompletionSpec(dirName, phase string) string {
	return `---
kit_metadata_version: 1
artifact: spec
feature:
  id: 1
  slug: v3
  dir: ` + dirName + `
workflow_version: 3
phase: ` + phase + `
---
# SPEC

## PURPOSE

Preserve consequential implementation rationale.

## CONTEXT

Native planning owns research and design.

## REQUIREMENTS

- The observable behavior remains correct.
- Non-goal: transcript ingestion.

## ACCEPTED PLAN

Implement the smallest coherent change and validate it.

## DECISIONS

- Accepted: keep semantic curation agent-owned because significance is contextual.

## DISCOVERIES

- Existing code and tests are the implementation evidence.

## VALIDATION

- ` + "`go test ./...`" + ` passed.

## OUTCOME

The requested behavior was implemented.

## REPOSITORY MEMORY

- Updated this specification with the final rationale and evidence.
`
}

func createFeatureTasks(t *testing.T, specsDir, dirName, tasks string) feature.Feature {
	t.Helper()
	featurePath := filepath.Join(specsDir, dirName)
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, "TASKS.md"), []byte(tasks), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	number, slug, ok := feature.ParseDirName(dirName)
	if !ok {
		t.Fatalf("ParseDirName(%q) failed", dirName)
	}
	return feature.Feature{
		Number:  number,
		Slug:    slug,
		DirName: dirName,
		Path:    featurePath,
		Phase:   feature.DeterminePhase(featurePath),
	}
}

func createFeatureFile(t *testing.T, specsDir, dirName, fileName, content string) {
	t.Helper()
	featurePath := filepath.Join(specsDir, dirName)
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, fileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
