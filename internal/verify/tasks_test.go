package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTaskBundlesParsesExecutableFields(t *testing.T) {
	dir := t.TempDir()
	tasksPath := filepath.Join(dir, "TASKS.md")
	content := `# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Parse declarations | todo | agent | |

## TASK LIST

- [ ] T001: Parse declarations [PLAN-01]

## TASK DETAILS

### T001

- **GOAL**: Parse task detail fields.
- **SCOPE**:
  - add parser
- **ACCEPTANCE**:
  - parser extracts commands
- **VERIFY**:
  - ` + "`go test ./...`" + `
- **EXPECTED FILES**:
  - ` + "`internal/verify/`" + `
- **RISK**: Medium; shared parser.
- **ROLLBACK**: Revert parser.
- **NOTES**: Keep it narrow.

## DEPENDENCIES

T001 first.

## NOTES

none
`
	if err := os.WriteFile(tasksPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	bundles, err := LoadTaskBundles(tasksPath, FeatureRefFromDir(dir), false)
	if err != nil {
		t.Fatalf("LoadTaskBundles() error = %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("len(bundles) = %d, want 1", len(bundles))
	}
	bundle := bundles[0]
	if bundle.TaskID != "T001" {
		t.Fatalf("TaskID = %q, want T001", bundle.TaskID)
	}
	if len(bundle.Verify) != 1 {
		t.Fatalf("len(Verify) = %d, want 1", len(bundle.Verify))
	}
	if got := bundle.Verify[0].Argv; len(got) != 3 || got[0] != "go" || got[1] != "test" || got[2] != "./..." {
		t.Fatalf("Argv = %#v, want go test ./...", got)
	}
	if len(bundle.ExpectedFiles) != 1 || bundle.ExpectedFiles[0] != "internal/verify/" {
		t.Fatalf("ExpectedFiles = %#v", bundle.ExpectedFiles)
	}
	if !bundle.HandoffNeeded {
		t.Fatal("expected medium-risk task to require handoff")
	}
}

func TestParseCommandRejectsShellSyntaxByDefault(t *testing.T) {
	_, err := ParseCommand("go test ./... && echo unsafe", "T001", 1, "TASKS.md", false)
	if err == nil {
		t.Fatal("expected shell syntax rejection")
	}
}

func TestParseCommandAllowsShellWhenExplicit(t *testing.T) {
	command, err := ParseCommand("echo ok && echo done", "T001", 1, "TASKS.md", true)
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	if !command.Shell {
		t.Fatal("expected shell command")
	}
}

func TestLoadTaskBundlesStopsLastTaskAtNextSection(t *testing.T) {
	dir := t.TempDir()
	tasksPath := filepath.Join(dir, "TASKS.md")
	content := `# TASKS

## TASK LIST

- [x] T001: Final task

## TASK DETAILS

### T001

- **NOTES**: Keep task notes scoped.

## DEPENDENCIES

- This belongs to the document, not T001.
`
	if err := os.WriteFile(tasksPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	bundles, err := LoadTaskBundles(tasksPath, FeatureRefFromDir(dir), false)
	if err != nil {
		t.Fatalf("LoadTaskBundles() error = %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("len(bundles) = %d, want 1", len(bundles))
	}
	if bundles[0].Notes != "Keep task notes scoped." {
		t.Fatalf("Notes = %q, want scoped task note only", bundles[0].Notes)
	}
}

func TestSelectExpectedFilesUsesTaskScope(t *testing.T) {
	bundles := []TaskBundle{
		{TaskID: "T001", ExpectedFiles: []string{"a.go"}, Verify: []Command{{Raw: "go test ./..."}}},
		{TaskID: "T002", ExpectedFiles: []string{"b.go", "a.go"}, Verify: []Command{{Raw: "go test ./..."}}},
		{TaskID: "T003", ExpectedFiles: []string{"unused.go"}},
	}

	expected := SelectExpectedFiles(bundles, "T002")
	if len(expected) != 2 || expected[0] != "b.go" || expected[1] != "a.go" {
		t.Fatalf("task expected files = %#v", expected)
	}

	expected = SelectExpectedFiles(bundles, "")
	if len(expected) != 2 || expected[0] != "a.go" || expected[1] != "b.go" {
		t.Fatalf("all expected files = %#v", expected)
	}
}
