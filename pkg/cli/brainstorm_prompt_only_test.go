package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestOutputExistingBrainstormPrompt_RegeneratesWithoutMutatingDocs(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	original := `# BRAINSTORM

## SUMMARY

Need better import validation.

## USER THESIS

Need better import validation for malformed CSV uploads.

## CODEBASE FINDINGS

findings

## AFFECTED FILES

files

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |

## QUESTIONS

questions

## OPTIONS

options

## RECOMMENDED STRATEGY

strategy

## NEXT STEP

kit spec sample
`
	writeFile(t, brainstormPath, original)

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	output := captureStdout(t, func() {
		err := outputExistingBrainstormPrompt([]string{"sample"}, projectRoot, cfg, true)
		if err != nil {
			t.Fatalf("outputExistingBrainstormPrompt() error = %v", err)
		}
	})

	if !strings.Contains(output, "Need better import validation for malformed CSV uploads.") {
		t.Fatalf("expected regenerated prompt to reuse thesis, got %q", output)
	}
	if !strings.Contains(output, filepath.Join(projectRoot, "docs", "notes", "0001-sample")) {
		t.Fatalf("expected regenerated prompt to mention feature notes directory, got %q", output)
	}

	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(content) != original {
		t.Fatalf("expected BRAINSTORM.md to remain unchanged")
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "docs", "notes", "0001-sample")); !os.IsNotExist(err) {
		t.Fatalf("expected --prompt-only to avoid creating notes directory, got %v", err)
	}
}

func TestOutputExistingBrainstormPrompt_SelectsExistingBrainstormFeature(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "BRAINSTORM.md"), `# BRAINSTORM

## SUMMARY

alpha summary

## USER THESIS

alpha thesis

## CODEBASE FINDINGS

findings

## AFFECTED FILES

files

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |

## QUESTIONS

questions

## OPTIONS

options

## RECOMMENDED STRATEGY

strategy

## NEXT STEP

kit spec alpha
`)
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0002-beta", "SPEC.md"), "# SPEC\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			err := outputExistingBrainstormPrompt(nil, projectRoot, cfg, true)
			if err != nil {
				t.Fatalf("outputExistingBrainstormPrompt() error = %v", err)
			}
		})
	})

	if !strings.Contains(output, "Select a feature to regenerate the brainstorm prompt for:") {
		t.Fatalf("expected selector prompt, got %q", output)
	}
	if !strings.Contains(output, "0001-alpha (brainstorm)") {
		t.Fatalf("expected brainstorm feature in selector, got %q", output)
	}
	if strings.Contains(output, "0002-beta") {
		t.Fatalf("expected selector to exclude non-brainstorm feature, got %q", output)
	}
	if !strings.Contains(output, "feature: **alpha**") {
		t.Fatalf("expected prompt for selected feature, got %q", output)
	}
}

func TestOutputExistingBrainstormPrompt_RejectsOutputFile(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "BRAINSTORM.md"), `# BRAINSTORM

## SUMMARY

summary

## USER THESIS

thesis

## CODEBASE FINDINGS

findings

## AFFECTED FILES

files

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |

## QUESTIONS

questions

## OPTIONS

options

## RECOMMENDED STRATEGY

strategy

## NEXT STEP

next
`)

	cfg := config.Default()
	previousOutput := brainstormOutput
	brainstormOutput = filepath.Join(projectRoot, "prompt.txt")
	defer func() {
		brainstormOutput = previousOutput
	}()

	err := outputExistingBrainstormPrompt([]string{"sample"}, projectRoot, cfg, true)
	if err == nil || !strings.Contains(err.Error(), "--prompt-only cannot be used with --output") {
		t.Fatalf("expected --output rejection, got %v", err)
	}
	if _, statErr := os.Stat(brainstormOutput); !os.IsNotExist(statErr) {
		t.Fatalf("expected prompt output file to remain absent, got %v", statErr)
	}
}
