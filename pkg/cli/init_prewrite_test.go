package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestInitDryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	home := setupInitHome(t)
	setWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "config.yml"), "blank_issues_enabled: false\n")
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "bug.yml"), "name: Bug\nbody: []\n")
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE.md"), "# legacy\n")
	writeFile(t, filepath.Join(root, makefilePath), "dev:\n")
	before := directorySnapshot(t, root)
	homeBefore := directorySnapshot(t, home)

	var output string
	withInitFlags(t, func() {
		initPlanDryRun = true
		output = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	for _, want := range []string{
		"Pre-write plan:",
		"create .kit.yaml",
		"skip .github/ISSUE_TEMPLATE/issue.md (opt-in; pass --issue-template)",
		"skip " + makefilePath,
		"Warning: existing issue template config: .github/ISSUE_TEMPLATE/config.yml",
		"Warning: existing issue form: .github/ISSUE_TEMPLATE/bug.yml",
		"Warning: existing issue template: .github/ISSUE_TEMPLATE.md",
		"Dry run complete. No files written.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q\n%s", want, output)
		}
	}
	if !sameSnapshot(before, directorySnapshot(t, root)) {
		t.Fatal("dry run changed the project directory")
	}
	if !sameSnapshot(homeBefore, directorySnapshot(t, home)) {
		t.Fatal("dry run changed the home directory")
	}
}

func TestInitIssueTemplateOptInAndGitignoreForce(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initIssueTemplate = true
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if got := readFile(t, filepath.Join(root, issueTemplatePath)); got != templates.IssueTemplate {
		t.Fatalf("template = %q", got)
	}
	cfg, err := config.Load(root)
	if err != nil || !cfg.Issues.IssueTemplateEnabled() {
		t.Fatalf("%v %#v", err, cfg)
	}

	ignored := t.TempDir()
	setWorkingDirectory(t, ignored)
	existing := "custom.log\n"
	writeFile(t, filepath.Join(ignored, gitignorePath), existing)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if got := readFile(t, filepath.Join(ignored, gitignorePath)); got != existing {
		t.Fatalf("non-interactive init changed gitignore: %q", got)
	}

	forced := t.TempDir()
	setWorkingDirectory(t, forced)
	writeFile(t, filepath.Join(forced, gitignorePath), existing)
	withInitFlags(t, func() {
		initOutputOnly = true
		initAllowOverwrite = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if got := readFile(t, filepath.Join(forced, gitignorePath)); !strings.Contains(got, "custom.log") || !strings.Contains(got, ".env") {
		t.Fatalf("force did not keep and extend gitignore: %q", got)
	}
}

func TestGatePrewriteConfirmAndReconcileDefault(t *testing.T) {
	merge := initRefreshFileChange{relativePath: gitignorePath, before: "a\n", after: "a\nb\n", result: instructionFileMerged}
	var out bytes.Buffer
	denied, _, err := gatePrewrite(t.TempDir(), []initRefreshFileChange{merge}, prewriteInit, true, false, false, strings.NewReader("n\n"), &out)
	if err != nil || denied[0].result != instructionFileSkipped || denied[0].after != "a\n" || !strings.Contains(out.String(), "Write .gitignore?") {
		t.Fatalf("%#v %v %s", denied, err, out.String())
	}
	allowed, _, err := gatePrewrite(t.TempDir(), []initRefreshFileChange{merge}, prewriteInit, true, false, false, strings.NewReader("y\n"), &out)
	if err != nil || allowed[0].result != instructionFileMerged {
		t.Fatalf("%#v %v", allowed, err)
	}
	kept, _, err := gatePrewrite(t.TempDir(), []initRefreshFileChange{merge}, prewriteReconcile, false, false, false, strings.NewReader(""), &out)
	if err != nil || kept[0].result != instructionFileMerged {
		t.Fatalf("reconcile overwrite %#v %v", kept, err)
	}

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "config.yml"), "blank_issues_enabled: false\n")
	create := initRefreshFileChange{relativePath: issueTemplatePath, after: templates.IssueTemplate, result: instructionFileCreated}
	skipped, notes, err := gatePrewrite(root, []initRefreshFileChange{create}, prewriteInit, false, false, true, strings.NewReader(""), &out)
	if err != nil || skipped[0].result != instructionFileSkipped || !strings.Contains(notes[issueTemplatePath], "--force") {
		t.Fatalf("%#v %v %v", skipped, notes, err)
	}
	forced, _, err := gatePrewrite(root, []initRefreshFileChange{create}, prewriteReconcile, false, true, false, strings.NewReader(""), &out)
	if err != nil || forced[0].result != instructionFileCreated {
		t.Fatalf("%#v %v", forced, err)
	}
}

func TestIssueTemplateSurfacesIncludeRootAndDocs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "ISSUE_TEMPLATE.md"), "# root\n")
	writeFile(t, filepath.Join(root, "docs", "ISSUE_TEMPLATE", "config.yml"), "blank_issues_enabled: false\n")
	var warnings []string
	for _, surface := range issueTemplateSurfaces(root) {
		warnings = append(warnings, surface.warning)
	}
	got := strings.Join(warnings, "\n")
	for _, want := range []string{"existing issue template: ISSUE_TEMPLATE.md", "existing issue template config: docs/ISSUE_TEMPLATE/config.yml"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	create := initRefreshFileChange{relativePath: issueTemplatePath, after: templates.IssueTemplate, result: instructionFileCreated}
	skipped, _, err := gatePrewrite(root, []initRefreshFileChange{create}, prewriteInit, false, false, true, strings.NewReader(""), &bytes.Buffer{})
	if err != nil || skipped[0].result != instructionFileSkipped {
		t.Fatalf("%#v %v", skipped, err)
	}
}

func TestReconcileDryRunKeepsInteractiveIssueTemplatePrompt(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "bug.yml"), "name: Bug\n")
	create := initRefreshFileChange{relativePath: issueTemplatePath, after: templates.IssueTemplate, result: instructionFileCreated}
	var out bytes.Buffer
	kept, notes, err := gatePrewrite(root, []initRefreshFileChange{create}, prewriteReconcile, true, false, true, strings.NewReader(""), &out)
	if err != nil || kept[0].result != instructionFileCreated || notes[issueTemplatePath] != "would ask before writing" || out.Len() != 0 {
		t.Fatalf("%#v %v %q", kept, err, notes)
	}
}

func TestInitLeavesPartialGlobalConfigUnlessForce(t *testing.T) {
	root := t.TempDir()
	home := setupInitHome(t)
	setWorkingDirectory(t, root)
	globalPath := filepath.Join(home, ".config", "kit", config.ConfigFileName)
	writeFile(t, globalPath, "schema_version: 2\n")
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if got := readFile(t, globalPath); got != "schema_version: 2\n" {
		t.Fatalf("global config changed without force: %q", got)
	}
	other := t.TempDir()
	setWorkingDirectory(t, other)
	withInitFlags(t, func() {
		initOutputOnly = true
		initAllowOverwrite = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if got := readFile(t, globalPath); !strings.Contains(got, "schema_version:") || got == "schema_version: 2\n" {
		t.Fatalf("force did not fill global config: %q", got)
	}
}

func TestOrgIssueTemplateWarning(t *testing.T) {
	previous := readmeCommandRunner
	t.Cleanup(func() { readmeCommandRunner = previous })
	readmeCommandRunner = prewriteCommandRunner{outputs: map[string]string{
		"git remote get-url origin":    "https://github.com/acme/app.git\n",
		"gh api users/acme --jq .type": "Organization\n",
	}}
	changes := []initRefreshFileChange{{relativePath: issueTemplatePath, result: instructionFileCreated}}
	got := orgIssueTemplateWarning(t.TempDir(), changes)
	if !strings.Contains(got, "acme") || !strings.Contains(strings.ToLower(got), "organization") || !strings.Contains(got, "does not download") {
		t.Fatal(got)
	}
	readmeCommandRunner = prewriteCommandRunner{}
	if !strings.Contains(orgIssueTemplateWarning(t.TempDir(), changes), "cannot detect") {
		t.Fatal("missing limitation")
	}
	if orgIssueTemplateWarning(t.TempDir(), nil) != "" {
		t.Fatal("warning without an issue template write")
	}
}

type prewriteCommandRunner struct {
	outputs map[string]string
}

func (r prewriteCommandRunner) Output(dir, name string, args ...string) ([]byte, error) {
	if out, ok := r.outputs[name+" "+strings.Join(args, " ")]; ok {
		return []byte(out), nil
	}
	return nil, errors.New(name)
}

func (r prewriteCommandRunner) OutputAllowError(dir, name string, args ...string) ([]byte, error) {
	return r.Output(dir, name, args...)
}

func sameSnapshot(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func directorySnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.IsDir() {
			snap[rel+"/"] = ""
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snap[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}
