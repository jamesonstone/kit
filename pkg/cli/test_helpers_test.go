package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func captureStdout(t *testing.T, function func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	function()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if _, err := io.Copy(&output, reader); err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func setupLifecycleTestProject(t *testing.T) (string, *config.Config) {
	t.Helper()
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), readyConstitutionForTest())
	for _, relativePath := range instructionArtifactPaths(cfg, instructionFileSelection{}, cfg.InstructionScaffoldVersion, true) {
		content, _, err := instructionArtifactContent(relativePath, cfg.InstructionScaffoldVersion)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(projectRoot, relativePath), content)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "docs", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return projectRoot, cfg
}

func readyConstitutionForTest() string {
	return `# CONSTITUTION

## PRINCIPLES

Clarity and correctness guide all project decisions.

## CONSTRAINTS

Keep generated and human-authored project guidance aligned.

## NON-GOALS

No additional non-goals for this test fixture.

## DEFINITIONS

Kit-managed project means a repository with Kit configuration and guidance.
`
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := document.Write(path, content); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func setWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist; stat error = %v", path, err)
	}
}

func restorePromptProfileState(t *testing.T, profile promptProfile, explicit bool) {
	t.Helper()
	previousProfile := selectedPromptProfile
	previousExplicit := selectedPromptProfileExplicit
	selectedPromptProfile = profile
	selectedPromptProfileExplicit = explicit
	t.Cleanup(func() {
		selectedPromptProfile = previousProfile
		selectedPromptProfileExplicit = previousExplicit
	})
}

func chdirForTest(t *testing.T, directory string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	}
}

func withStdin(t *testing.T, input string, function func() string) string {
	t.Helper()
	previous := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = reader
	defer func() {
		os.Stdin = previous
		_ = reader.Close()
	}()
	return function()
}

func withClipboardCopy(t *testing.T, copyFunction func(string) error) {
	t.Helper()
	previous := clipboardCopyFunc
	clipboardCopyFunc = copyFunction
	t.Cleanup(func() { clipboardCopyFunc = previous })
}

func validV2SpecWithPhase(dirName, phase string) string {
	status, confidence, unresolved := "open", 0, 1
	switch phase {
	case "ready", "implement", "validate", "reflect", "deliver", "complete":
		status, confidence, unresolved = "ready", 95, 0
	case "blocked":
		status = "blocked"
	}
	id, slug, _ := strings.Cut(dirName, "-")
	return `---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: ` + phase + `
clarification:
  status: ` + status + `
  confidence: ` + strconv.Itoa(confidence) + `
  unresolved_questions: ` + strconv.Itoa(unresolved) + `
feature:
  id: "` + id + `"
  slug: ` + slug + `
  dir: ` + dirName + `
---
# SPEC

## THESIS

Thesis.

## CONTEXT

Context.

## CLARIFICATIONS

None.

## REQUIREMENTS

- Requirement.

## ASSUMPTIONS

None.

## ACCEPTANCE CRITERIA

- AC-001: Criterion.

## IMPLEMENTATION PLAN

Plan.

## TASK CHECKLIST

- [x] Task.

## VALIDATION MAP

- AC-001 -> test.

## REFLECTION NOTES

None.

## DOCUMENTATION UPDATES

Current.

## DELIVERY DECISION

None.

## EVIDENCE

Evidence.
`
}
