package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltInDynamicPromptsDeclareContextRequirements(t *testing.T) {
	dynamic := map[string]bool{
		"kit spec":          true,
		"workflow spec":     true,
		"support resume":    true,
		"support reconcile": true,
		"support dispatch":  true,
		"skill mine":        true,
	}

	for _, prompt := range builtInKitPromptSource().Prompts {
		command := prompt.Identity.CommandName()
		if !dynamic[command] {
			continue
		}
		if len(prompt.ContextRequirements) == 0 {
			t.Fatalf("dynamic prompt %q has no context requirements", command)
		}
		delete(dynamic, command)
	}

	if len(dynamic) != 0 {
		t.Fatalf("dynamic prompt checks did not run for %v", dynamic)
	}
}

func TestActivePromptFeatureContextInfersNewestActiveFeature(t *testing.T) {
	projectRoot := newPromptContextProject(t)
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writePromptContextFeatureDocs(t, featurePath, true)
	setWorkingDirectory(t, projectRoot)

	ctx, err := activePromptFeatureContext("workflow spec", "SPEC.md")
	if err != nil {
		t.Fatalf("activePromptFeatureContext() error = %v", err)
	}
	if same, err := samePromptContextPath(ctx.ProjectRoot, projectRoot); err != nil || !same {
		t.Fatalf("ProjectRoot = %q, want %q", ctx.ProjectRoot, projectRoot)
	}
	if ctx.Feature.Slug != "alpha" {
		t.Fatalf("Feature.Slug = %q, want alpha", ctx.Feature.Slug)
	}
}

func TestActivePromptFeatureContextCapturesMissingFeatureThroughEditor(t *testing.T) {
	projectRoot := newPromptContextProject(t)
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writePromptContextFeatureDocs(t, featurePath, true)
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), "- [x] done\n<!-- REFLECTION_COMPLETE -->\n")
	setWorkingDirectory(t, projectRoot)
	stubPromptEditorInput(t, "feature: alpha", true)

	var ctx *promptFeatureContext
	output := withStdin(t, "y\n", func() string {
		var err error
		ctx, err = activePromptFeatureContext("workflow spec", "SPEC.md")
		if err != nil {
			t.Fatalf("activePromptFeatureContext() error = %v", err)
		}
		return ""
	})

	if output != "" {
		t.Fatalf("unexpected withStdin return = %q", output)
	}
	if ctx.Feature.Slug != "alpha" {
		t.Fatalf("Feature.Slug = %q, want alpha", ctx.Feature.Slug)
	}
}

func TestCollectMissingPromptContextDeclineFails(t *testing.T) {
	output := withStdin(t, "n\n", func() string {
		_, err := collectMissingPromptContext(
			"support dispatch",
			"a task list",
			"dispatch tasks",
			newFreeTextInputConfig(true, "", false, true),
		)
		if err == nil {
			t.Fatalf("expected collectMissingPromptContext() to fail")
		}
		if !strings.Contains(err.Error(), "requires a task list") {
			t.Fatalf("unexpected error = %v", err)
		}
		return ""
	})
	if output != "" {
		t.Fatalf("unexpected withStdin return = %q", output)
	}
}

func TestCollectMissingPromptContextEmptyEditorContentFails(t *testing.T) {
	stubPromptEditorInput(t, "   ", true)

	_ = withStdin(t, "y\n", func() string {
		_, err := collectMissingPromptContext(
			"support dispatch",
			"a task list",
			"dispatch tasks",
			newFreeTextInputConfig(true, "", false, true),
		)
		if err == nil {
			t.Fatalf("expected collectMissingPromptContext() to fail")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Fatalf("unexpected error = %v", err)
		}
		return ""
	})
}
func TestRenderSupportDispatchPromptUsesEditorContext(t *testing.T) {
	setWorkingDirectory(t, t.TempDir())
	stubPromptEditorInput(t, "- Update prompt command\n- Add prompt list", true)

	_ = withStdin(t, "y\n", func() string {
		rendered, err := renderSupportDispatchPrompt()
		if err != nil {
			t.Fatalf("renderSupportDispatchPrompt() error = %v", err)
		}
		if !strings.Contains(rendered, "Prepare an Agent Team Plan") {
			t.Fatalf("expected dispatch prompt, got %q", rendered)
		}
		if !strings.Contains(rendered, "D001") || !strings.Contains(rendered, "D002") {
			t.Fatalf("expected normalized dispatch tasks, got %q", rendered)
		}
		return ""
	})
}

func TestRenderWorkflowSpecPromptInfersActiveFeature(t *testing.T) {
	projectRoot := newPromptContextProject(t)
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writePromptContextFeatureDocs(t, featurePath, true)
	setWorkingDirectory(t, projectRoot)

	rendered, err := renderWorkflowSpecPrompt()
	if err != nil {
		t.Fatalf("renderWorkflowSpecPrompt() error = %v", err)
	}
	assertV2SpecPromptContract(t, rendered)
	assertV2SpecPromptExcludesV1StageAssumptions(t, rendered)
	if !strings.Contains(rendered, "Supervise feature `alpha`") {
		t.Fatalf("expected feature slug in v2 specification prompt, got %q", rendered)
	}
	if strings.Contains(rendered, "## Subagent Orchestration") {
		t.Fatalf("workflow spec prompt should not append generic subagent guidance, got %q", rendered)
	}
}

func stubPromptEditorInput(t *testing.T, text string, changed bool) {
	t.Helper()

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	t.Cleanup(func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	})

	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		return text, changed, nil
	}
}

func newPromptContextProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT_PROGRESS_SUMMARY\n")
	return projectRoot
}

func writePromptContextFeatureDocs(t *testing.T, featurePath string, includeBrainstorm bool) {
	t.Helper()

	if includeBrainstorm {
		writeFile(t, filepath.Join(featurePath, "BRAINSTORM.md"), "# BRAINSTORM\n\n## SUMMARY\n\nalpha brainstorm\n")
	}
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), "# SPEC\n\n## SUMMARY\n\nalpha summary\n")
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), "# PLAN\n")
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), "- [ ] implement alpha\n")
}

func samePromptContextPath(left, right string) (bool, error) {
	normalizedLeft, err := filepath.EvalSymlinks(left)
	if err != nil {
		return false, err
	}
	normalizedRight, err := filepath.EvalSymlinks(right)
	if err != nil {
		return false, err
	}
	return normalizedLeft == normalizedRight, nil
}
